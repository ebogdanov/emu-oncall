package grafana

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/ebogdanov/emu-oncall/internal/metrics"

	"github.com/rs/zerolog"

	"github.com/ebogdanov/emu-oncall/internal/user"

	"github.com/ebogdanov/emu-oncall/internal/config"
	"github.com/ebogdanov/emu-oncall/plugin"
)

const (
	endpointIncidentDetails = "%s/api/plugin-proxy/grafana-oncall-ui/api/internal/v1/alertgroups?search=%d"
)

const (
	waitTime         = 100 * time.Millisecond
	oncallSyncPeriod = 1 * time.Minute
	httpTimeout      = 15 * time.Second
)

const (
	titleDuty         = "-= Duty Status =-"
	templateStartDuty = "Your duty %s started"
	templateEndDuty   = "Duty %s is over"
)

// Lookup for something like " incident #123 #
var (
	regexpIncidentID = regexp.MustCompile(`\sincident\s#(\d+)\s`)
)

var (
	errNotAssigneeSchedule = errors.New("not found assignee for schedule")
	errEmptyOnCallURL      = errors.New("grafana.oncall.url is empty. Skip load incident details")
	errNotFoundIncident    = errors.New("incident is not found")
)

type Service interface {
	IncidentDetails(context.Context, int) (Render, error)
	Start(context.Context)
}

type ScheduleProcessor interface {
	Current(ctx context.Context) (*ScheduleItem, error)
}

type Instance struct {
	cfg           *config.Grafana
	p             plugin.Plugin
	repoUser      *user.Storage
	client        httpClient
	logger        zerolog.Logger
	notifyChannel chan *scheduleEvent
	sm            *sync.RWMutex
	duty          *OnDuty
	promMetrics   *metrics.Storage
}

func New(cfg *config.Grafana, p plugin.Plugin, u *user.Storage, l zerolog.Logger, pm *metrics.Storage) Service {
	auth := &engineToken{Token: cfg.Token}

	instance := &Instance{
		p:             p,
		repoUser:      u,
		cfg:           cfg,
		sm:            &sync.RWMutex{},
		notifyChannel: make(chan *scheduleEvent, 5000),
		client:        newAPIClient(auth.HeaderValue(), pm),
		logger:        l.With().Str("component", "grafana").Logger(),
		duty:          NewOnDuty(),
		promMetrics:   pm,
	}
	return instance
}

func (g *Instance) IncidentDetails(ctx context.Context, incidentID int) (Render, error) {
	var (
		alertGroup alertGroupResponse
	)

	if !g.cfg.IncidentDetails || incidentID == 0 {
		return nil, nil
	}

	if g.cfg.URL == "" {
		return nil, errEmptyOnCallURL
	}

	requestURL := fmt.Sprintf(endpointIncidentDetails,
		strings.TrimSuffix(g.cfg.URL, "/"), incidentID)

	resp, err := g.client.Get(ctx, requestURL)

	if err != nil {
		return nil, errors.Join(err, errors.New("Request URL: "+requestURL))
	}

	// Try to check if this is response with JSON error
	err = json.Unmarshal(resp, &alertGroup)
	if err != nil {
		return nil, err
	}

	for i := range alertGroup.Results {
		if alertGroup.Results[i].InsideOrganizationNumber == incidentID {
			return &alertGroup.Results[i], nil
		}
	}

	err = errNotFoundIncident
	if alertGroup.Message != "" {
		err = errors.New(alertGroup.Message)
	}

	if alertGroup.Detail != "" {
		err = errors.New(alertGroup.Detail)
	}

	return nil, err
}

// checkOnCall will download file, parse it and return current user oncall now
func (g *Instance) checkOnCall(ctx context.Context, schedule *config.ScheduleEntry) (*ScheduleItem, error) {
	iCallURL := strings.TrimSuffix(g.cfg.OnCall.URL, "/") + schedule.IcalURL

	client := NewIcsInstance(iCallURL, g.promMetrics)
	current, err := client.Current(ctx)

	if err != nil || current == nil {
		return nil, err
	}

	if current.Name == "" {
		current.Name = schedule.Name
	}

	current.transport = schedule.Transport
	current.callbackURL = schedule.CallbackURL

	return current, nil
}

func (g *Instance) Start(ctx context.Context) {
	scheduleTicker := time.NewTicker(oncallSyncPeriod)

	for {
		select {
		// Every 2 minutes, load & checks schedules
		case <-scheduleTicker.C:
			func() {
				g.checkSchedules(ctx)
			}()

		// Notification channel
		case msg := <-g.notifyChannel:
			go func(ctx context.Context, notifyMsg *scheduleEvent) {
				g.notify(ctx, notifyMsg)
			}(ctx, msg)

		case <-ctx.Done():
			close(g.notifyChannel)
			scheduleTicker.Stop()
			return

		default:
			time.Sleep(waitTime)
		}
	}
}

func (g *Instance) notify(ctx context.Context, msg *scheduleEvent) {
	//
	var (
		err       error
		res       []byte
		recipient *user.List
		body      []byte
	)

	switch msg.Transport {
	case callback:
		req := &formattedAlert{
			Title:   titleDuty,
			Message: msg.Msg,
		}

		if body, err = json.Marshal(req); err == nil {
			res, err = g.client.Post(ctx, msg.URL, string(body))
		}
	default:
		recipient, err = g.repoUser.DBQuery(ctx, user.Options{UserID: msg.UserID})

		msgID := fmt.Sprintf("duty_start_%s", msg.ScheduleName)

		if err == nil && len(recipient.Result) > 0 {
			err = g.p.SendSms(ctx, recipient.Result[0], msgID, msg.Msg)
		}
	}

	if err != nil {
		g.logger.Error().
			Err(err).
			Msg("Unable to send notification")

		return
	}

	g.logger.Debug().
		Bytes("result", res).
		Interface("msg", msg).
		Msg("Notification was sent")
}

func (g *Instance) checkSchedules(ctx context.Context) {
	g.sm.Lock()
	defer g.sm.Unlock()

	for _, item := range g.cfg.Schedules {
		ctxT, cancelFunc := context.WithTimeout(ctx, 10*time.Second)

		g.logger.Debug().
			Str("schedule", item.Name).
			Str("url", item.IcalURL).
			Msg("download ICS file from URL")

		schedule, err := g.checkOnCall(ctxT, item)
		cancelFunc()

		if err != nil {
			g.promMetrics.SchedulesCounter.WithLabelValues(item.Name, "error").Inc()

			g.logger.Error().
				Str("schedule", item.Name).
				Err(err).
				Msg("failed to get entries for schedule")

			continue
		}

		g.promMetrics.SchedulesCounter.WithLabelValues(schedule.Name, "ok").Inc()
		g.scheduleEntry(schedule)
	}
}

func (g *Instance) scheduleEntry(schedule *ScheduleItem) {
	name := schedule.Name

	curr := g.duty.Get(name)

	currentOnCall, err := curr.Peek()
	// This is first record, just add it
	if err != nil {
		curr.Push(schedule.Users)
		return
	}

	startDuty, endDuty := compareSlices(schedule.Users, currentOnCall)
	if len(startDuty)+len(endDuty) == 0 {
		return
	}

	// Hey, your duty is starting
	for _, userID := range startDuty {
		g.promMetrics.Notifications.WithLabelValues("duty_start").Inc()

		g.logger.Debug().
			Str("username", userID).
			Str("schedule", name).
			Msg("send start duty notification")

		text := fmt.Sprintf(templateStartDuty, name)

		g.notifyChannel <- &scheduleEvent{
			Transport: schedule.transport,
			UserID:    userID,
			Msg:       text,
			Title:     titleDuty,
			URL:       schedule.callbackURL}
	}

	// What a pity - duty is over :(
	for _, userID := range endDuty {
		g.promMetrics.Notifications.WithLabelValues("duty_end").Inc()

		g.logger.Debug().
			Str("username", userID).
			Str("schedule", name).
			Msg("send duty end notification")

		text := fmt.Sprintf(templateEndDuty, name)
		g.notifyChannel <- &scheduleEvent{
			Transport: schedule.transport,
			UserID:    userID,
			Msg:       text,
			Title:     titleDuty,
			URL:       schedule.callbackURL}
	}

	_, _ = curr.Pop()
	curr.Push(schedule.Users)
}

// nolint:gocritic
func compareSlices(a, b []string) ([]string, []string) {
	cntA := make(map[string]int)
	cntB := make(map[string]int)

	for _, item := range a {
		cntA[item]++
	}

	for _, item := range b {
		cntB[item]++
	}

	var (
		diffA []string
		diffB []string
	)
	for item, countA := range cntA {
		countB, exists := cntB[item]
		if !exists || countA != countB {
			for i := 0; i < countA; i++ {
				diffA = append(diffA, item)
			}
		}
	}

	for item, countB := range cntB {
		countA, exists := cntA[item]
		if !exists || countA != countB {
			for i := 0; i < countB; i++ {
				diffB = append(diffB, item)
			}
		}
	}

	return diffA, diffB
}
