package grafana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/ebogdanov/emu-oncall/internal/user"

	"github.com/ebogdanov/emu-oncall/internal/config"
	"github.com/ebogdanov/emu-oncall/internal/logger"

	"github.com/ebogdanov/emu-oncall/plugin"
)

const (
	endpointGrafanaUsers    = "%s/api/users?perpage=1000&page=%d"
	endpointOnCallSchedules = "%s/api/v1/schedules/"
	endpointIncidentDetails = "%s/api/plugin-proxy/grafana-oncall-app/api/internal/v1/alertgroups?search=%d"
)

const (
	waitTime         = 100 * time.Millisecond
	engineSyncPeriod = 1 * time.Minute
	oncallSyncPeriod = 2 * time.Minute
	httpTimeout      = 15 * time.Second
)

const (
	titleStartDuty    = `Начало смены`
	templateStartDuty = `Произошла ротация дежурного на смене %s`
	templateEndDuty   = `Ваше дежурство на смене %s закончилось`
)

var (
	regexpIncidentID = regexp.MustCompile(`incident\s#(\d+)\s`)
)

type Service interface {
	CanLoadIncident() bool
	IncidentDetails(context.Context, int) Render
	CheckSchedules(context.Context) bool
	Start(context.Context)
}

type scheduleCache = map[string]map[string]bool

type Instance struct {
	cfg           *config.Grafana
	notify        plugin.Plugin
	repoUser      *user.Storage
	httpClient    *http.Client
	logger        *logger.Instance
	onCallNow     scheduleCache
	notifyChannel chan *scheduleEvent
	engineAuth    authCredentials
	m             *sync.RWMutex
	onceTime      *sync.Once
}

func New(cfg *config.Grafana, p plugin.Plugin, u *user.Storage, l *logger.Instance) Service {
	instance := &Instance{
		notify:        p,
		repoUser:      u,
		cfg:           cfg,
		m:             &sync.RWMutex{},
		onceTime:      &sync.Once{},
		notifyChannel: make(chan *scheduleEvent),
		engineAuth: &engineToken{User: cfg.User,
			Password: cfg.Password,
			Token:    cfg.Token},
		logger: l,
	}
	return instance
}

func (g *Instance) IncidentDetails(ctx context.Context, incidentID int) Render {
	var (
		alertGroupResult alertGroupResponse
	)

	if incidentID == 0 {
		return nil
	}

	if g.cfg.OnCallURL == "" {
		g.logger.Debug().
			Str("component", "grafana").
			Msg("grafana.oncall.url is empty. Skip load schedules")

		return nil
	}

	requestURL := fmt.Sprintf(endpointIncidentDetails,
		strings.TrimSuffix(g.cfg.OnCallURL, "/"), incidentID)

	token := g.engineAuth.HeaderValue()
	resp, err := g.apiCall(ctx, requestURL, token, "")

	if err != nil {
		g.logger.Error().
			Str("component", "grafana").
			Err(err).
			Str("url", requestURL).
			Msg("Unable to load incident from Grafana OnCall")

		return nil
	}

	g.logger.Info().
		Str("component", "grafana").
		Bytes("response", resp).
		Int("incidentID", incidentID).
		Msg("Loaded incident details from Grafana API")

	// Try to check if this is response with JSON error
	err = json.Unmarshal(resp, &alertGroupResult)
	if err != nil {
		g.logger.Error().
			Str("component", "grafana").
			Err(err).
			Msg("Unable to parse response JSON from Grafana OnCall")

		return nil
	}

	for i := range alertGroupResult.Results {
		if alertGroupResult.Results[i].InsideOrganizationNumber == incidentID {
			return &alertGroupResult.Results[i]
		}
	}

	return nil
}

func (g *Instance) CanLoadIncident() bool {
	return g.cfg.TrackIncidentDetails
}

func (g *Instance) CheckSchedules(_ context.Context) bool {
	var (
		scheduleList schedulesResponse
	)

	// Read ICal file, if possible
	for i := range scheduleList.Results {
		g.scheduleEntry(scheduleList.Results[i])
	}

	return true
}

func (g *Instance) Start(ctx context.Context) {
	g.initHTTPClient()

	g.CheckSchedules(ctx)

	go func() {
		for {
			select {
			// Every 1 minute, load schedules
			case <-time.After(oncallSyncPeriod):
				if g.cfg.TrackSchedules {
					g.CheckSchedules(ctx)
				}

			// Notification channel
			case msg := <-g.notifyChannel:
				go func(ctx context.Context, notifyMsg *scheduleEvent) {
					g.sendNotification(ctx, notifyMsg)
				}(ctx, msg)

			case <-ctx.Done():
				close(g.notifyChannel)
				return

			default:
				time.Sleep(waitTime)
			}
		}
	}()
}

func (g *Instance) sendNotification(ctx context.Context, msg *scheduleEvent) {
	//
	var (
		err error
		res []byte
	)

	switch msg.Transport {
	case push:
		var recipient *user.List
		recipient, err = g.repoUser.DBQuery(ctx, user.Options{UserID: msg.UserID})

		if err == nil && len(recipient.Result) > 0 {
			err = g.notify.SendSms(ctx, recipient.Result[0], "", msg.Msg)
		}

	case callback:
		var body []byte
		req := &formattedAlert{
			Title:   titleStartDuty,
			Message: msg.Msg,
		}

		if body, err = json.Marshal(req); err == nil {
			res, err = g.apiCall(ctx, "http://localhost:8080/integrations/v1/formatted_webhook/E06HHKeTcRiyTjJ13ChH4SLDd/", "", string(body))
		}
	}

	g.logger.Debug().
		Err(err).
		Bytes("result", res).
		Interface("msg", msg).
		Send()
}

func (g *Instance) initHTTPClient() *http.Client {
	if g.httpClient != nil {
		return g.httpClient
	}

	httpClient := &http.Client{Timeout: httpTimeout}

	g.httpClient = httpClient

	return httpClient
}

func (g *Instance) apiCall(ctx context.Context, apiURL, token, postBody string) ([]byte, error) {
	var (
		req *http.Request
		err error
	)

	// Make call
	if postBody == "" {
		req, err = http.NewRequest(http.MethodGet, apiURL, http.NoBody)
	} else {
		req, err = http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer([]byte(postBody)))
	}

	req = req.WithContext(ctx)
	req.Header.Add("Content-Type", "application/json")

	if token != "" {
		req.Header.Add("Authorization", token)
	}

	if err != nil {
		return nil, err
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()

	return bodyBytes, err
}

func (g *Instance) scheduleEntry(item scheduleItem) {
	g.m.RLock()
	if len(item.OnCallNow) == 0 {
		g.logger.Error().
			Str("name", item.Name).
			Msg("Schedule is empty")

		return
	}

	scheduleName := item.Name
	currentSchedule := g.onCallNow
	newSchedule := make(scheduleCache)

	currentItem, sendNotification := currentSchedule[scheduleName]
	g.m.RUnlock()

	for _, userID := range item.OnCallNow {
		// Process cache items
		// 1. Check if item is in cache
		// 2. If not – add
		// 3. If yes – check users, if not match – send notification

		if !sendNotification {
			if _, sendNotification = currentItem[userID]; sendNotification {
				continue
			}
		}

		if newSchedule[scheduleName] == nil {
			newSchedule[scheduleName] = make(map[string]bool)
		}
		newSchedule[scheduleName][userID] = true

		if sendNotification {
			go func(userId, scheduleName string) {
				text := fmt.Sprintf(templateStartDuty, scheduleName)
				g.notifyChannel <- &scheduleEvent{
					Transport: callback, UserID: userId, Msg: text, Title: titleStartDuty}
			}(userID, scheduleName)
		}
	}

	g.m.Lock()
	g.onCallNow = newSchedule
	g.m.Unlock()

	text := fmt.Sprintf(templateEndDuty, scheduleName)
	for name, currentItem := range currentSchedule {
		// Schedule is disabled or deleted for some reason
		if _, ok := newSchedule[name]; !ok {
			for uid := range currentItem {
				go func(userId, scheduleName string) {
					g.notifyChannel <- &scheduleEvent{
						Transport: push, UserID: userId, Msg: text, ScheduleName: scheduleName}
				}(uid, name)
			}

			continue
		}

		for uid := range currentItem {
			if _, ok := newSchedule[name][uid]; !ok {
				go func(userId, scheduleName string) {
					g.notifyChannel <- &scheduleEvent{
						Transport: push, UserID: userId, Msg: text, ScheduleName: scheduleName}
				}(uid, name)
			}
		}
	}
}
