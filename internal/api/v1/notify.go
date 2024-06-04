package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/ebogdanov/emu-oncall/internal/events"
	grafana2 "github.com/ebogdanov/emu-oncall/internal/grafana"
	user2 "github.com/ebogdanov/emu-oncall/internal/user"

	"strings"
	"sync"
	"time"

	"errors"

	"github.com/ebogdanov/emu-oncall/internal/logger"
	"github.com/ebogdanov/emu-oncall/plugin"
)

var (
	errUserNotFound         = errors.New("user-not-found")
	errUserPhoneNotVerified = errors.New("phonenumber-not-verified")
	errInvalidEmail         = errors.New("invalid-email")
	errEmptyMessage         = errors.New("empty-message-text")
	errInvalidPushToken     = errors.New("push-token-not-found")
)

const (
	sentSuccess = "Sent"
	callSuccess = "Was called"
)

const (
	notifyBySms   = "sms"
	notifyByPhone = "phone"
)

type NotifyResponse struct {
	Error string `json:"error,omitempty"`
}

type Notify struct {
	plugin     plugin.Plugin
	logger     *logger.Instance
	userData   user2.Storage
	actionsLog events.Service
	grafanaSvc grafana2.Service
	cache      sync.Map
}

type NotifyRequest struct {
	Email   string
	Message string
}

func NewNotify(dbData *user2.Storage, pl plugin.Plugin, log *logger.Instance, actionLogger events.Service, grafanaService grafana2.Service) *Notify {
	return &Notify{
		userData:   *dbData,
		plugin:     pl,
		cache:      sync.Map{},
		logger:     log,
		actionsLog: actionLogger,
		grafanaSvc: grafanaService,
	}
}

func (nh *Notify) ServeHTTP(response http.ResponseWriter, req *http.Request) {
	var (
		result interface{}
		err    error
		resp   []byte
	)

	if strings.HasPrefix(req.RequestURI, "/make_call") {
		result, err = nh.PhoneCall(req.Context(), req)
	} else {
		result, err = nh.SMS(req.Context(), req)
	}

	if err == nil {
		resp, err = json.Marshal(result)
		if err == nil {
			response.WriteHeader(http.StatusOK)
			_, _ = response.Write(resp)

			return
		}
	}

	response.WriteHeader(http.StatusInternalServerError)
	resp, _ = json.Marshal(err.Error())

	_, _ = response.Write(resp)
}

func (nh *Notify) PhoneCall(ctx context.Context, req *http.Request) (interface{}, error) {
	// Make select query to DB
	// if error - lookup in local cache
	// email=admin%40localhost&
	// message=You+are+invited+to+check+an+incident+from+Grafana+OnCall.+Alert+via+Formatted+Webhook+%3Ablush%3A+with+title+TestAlert%3A+The+whole+system+is+down+triggered+1+times

	return nh.notification(ctx, req, notifyByPhone)
}

func (nh *Notify) SMS(ctx context.Context, req *http.Request) (interface{}, error) {
	return nh.notification(ctx, req, notifyBySms)
}

func (nh *Notify) notification(ctx context.Context, req *http.Request, notificationType string) (interface{}, error) {
	request, err := nh.validate(ctx, req, notifyBySms)
	if err != nil {
		nh.logger.Error().
			Str("component", "notify").
			Err(err).
			Interface("notification_type", notificationType).
			Interface("email", request.Email).
			Msg("unable to to validate notification request")

		return &NotifyResponse{Error: err.Error()}, nil
	}

	userResult, err := nh.user(ctx, request.Email)
	if err != nil || userResult == nil {
		nh.logger.Error().
			Err(err).
			Interface("notification_type", notificationType).
			Interface("email", request.Email).
			Msg("Unable to lookup user to send notification")

		return &NotifyResponse{Error: errUserNotFound.Error()}, nil
	}

	if !userResult.IsPhoneNumberVerified || userResult.PhoneNumber == "" {
		return &NotifyResponse{Error: errUserPhoneNotVerified.Error()}, nil
	}

	nh.logger.Info().
		Interface("user", userResult).
		Interface("message", request).
		Str("notification_type", notificationType).
		Msg("Sending to user")

	alertID := ""
	msg := request.Message
	incidentID := grafana2.IncidentID(msg)

	// Load more details
	if nh.grafanaSvc.CanLoadIncident() {
		// Get details
		alertGroup := nh.grafanaSvc.IncidentDetails(ctx, incidentID)

		if alertGroup != nil {
			alertID = alertGroup.AlertID()

			if notificationType == notifyByPhone {
				// Render message
				msg = alertGroup.PhoneCall(msg)
			}

			if notificationType == notifyBySms {
				// Render message
				msg = alertGroup.SMSText(msg)
			}
		}
	}

	if notificationType == notifyByPhone {
		err = nh.plugin.CallPhone(ctx, *userResult, alertID, msg)
	} else {
		err = nh.plugin.SendSms(ctx, *userResult, alertID, msg)
	}

	if err != nil {
		nh.logger.Error().
			Err(err).
			Interface("email", request.Email).
			Interface("phone", userResult.PhoneNumber).
			Str("notification_type", notificationType).
			Int("incident_id", incidentID).
			Str("alert_id", alertID).
			Msg("Unable to Send notification")

		if err == sql.ErrNoRows {
			err = errInvalidPushToken
		}

		nh.logAction(userResult.ID,
			userResult.PhoneNumber, notificationType, false, err.Error())

		return &NotifyResponse{Error: err.Error()}, nil
	}

	successStr := sentSuccess
	if notificationType == notifyByPhone {
		successStr = callSuccess
	}

	nh.logger.Info().
		Interface("notification_type", notificationType).
		Interface("phone", userResult.PhoneNumber).
		Msg(successStr)

	nh.logAction(userResult.ID, userResult.PhoneNumber, notificationType, true, successStr)

	nh.cache.Store(request.Email, *userResult)

	return &NotifyResponse{}, nil
}

func (nh *Notify) validate(_ context.Context, req *http.Request, _ string) (*NotifyRequest, error) {
	formData := req.PostForm

	nh.logger.Info().
		Interface("request", formData).
		Str("url", req.URL.String()).
		Msg("Incoming request")

	email := formData.Get("email")
	if !nh.isValidEmail(email) {
		return nil, errInvalidEmail
	}

	alertText := formData.Get("message")

	if alertText == "" {
		return nil, errEmptyMessage
	}

	return &NotifyRequest{Email: email, Message: alertText}, nil
}

//nolint:interfacer
func (nh *Notify) user(ctx context.Context, email string) (*user2.Item, error) {
	// Lookup in DB, if not found - use user from local cache

	userResult, err := nh.userData.ByEmail(ctx, email)
	if err == nil {
		return userResult, nil
	}

	if cachedItem, ok := nh.cache.Load(email); ok {
		return cachedItem.(*user2.Item), nil
	}

	return nil, errUserNotFound
}

func (nh *Notify) isValidEmail(email string) bool {
	return strings.Contains(email, "@")
}

func (nh *Notify) logAction(userID, phoneNumber, route string, success bool, msg string) {
	if nh.actionsLog == nil {
		return
	}

	record := &events.Item{
		Timestamp: time.Now(),
		UserID:    userID,
		Recipient: phoneNumber,
		Channel:   route,
		Success:   success,
		Msg:       msg,
	}

	_ = nh.actionsLog.Add(record)
}
