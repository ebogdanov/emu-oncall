package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/ebogdanov/emu-oncall/internal/metrics"

	"github.com/ebogdanov/emu-oncall/internal/events"
	"github.com/ebogdanov/emu-oncall/internal/grafana"
	"github.com/ebogdanov/emu-oncall/internal/user"

	"strings"
	"sync"
	"time"

	"errors"

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
	smsTextNotify   = "sms"
	phoneCallNotify = "phone"
)

type notifyResponse struct {
	Error string `json:"error,omitempty"`
}

type Notify struct {
	plugin     plugin.Plugin
	logger     zerolog.Logger
	users      user.Storage
	actionsLog events.Service
	grafanaSvc grafana.Service
	cache      sync.Map
	pm         *metrics.Storage
}

type NotifyRequest struct {
	Email   string
	Message string
}

func NewNotify(dbData *user.Storage, pl plugin.Plugin, logger zerolog.Logger, actionLogger events.Service, gfSvc grafana.Service, promMetrics *metrics.Storage) *Notify {
	return &Notify{
		users:      *dbData,
		plugin:     pl,
		cache:      sync.Map{},
		logger:     logger.With().Str("component", "notify").Logger(),
		actionsLog: actionLogger,
		grafanaSvc: gfSvc,
		pm:         promMetrics,
	}
}

func (n *Notify) ServeHTTP(response http.ResponseWriter, req *http.Request) {
	n.pm.Notifications.WithLabelValues("total").Inc()

	notifyType := smsTextNotify
	if strings.Contains(req.RequestURI, "/make_call") {
		notifyType = phoneCallNotify
	}

	n.pm.Notifications.WithLabelValues(notifyType).Inc()

	result, err := n.notification(req.Context(), req, notifyType)
	if err == nil {
		resp, _ := json.Marshal(result)

		n.pm.Notifications.WithLabelValues("success").Inc()
		response.WriteHeader(http.StatusOK)
		_, _ = response.Write(resp)

		return
	}

	n.pm.Notifications.WithLabelValues("error_" + notifyType).Inc()
	n.pm.Notifications.WithLabelValues("fail").Inc()
	resp, _ := json.Marshal(&notifyResponse{Error: err.Error()})

	response.WriteHeader(http.StatusInternalServerError)
	_, _ = response.Write(resp)
}

func (n *Notify) notification(ctx context.Context, req *http.Request, notificationType string) (interface{}, error) {
	request, err := n.validate(ctx, req, notificationType)

	if err != nil {
		n.logger.Error().
			Err(err).
			Interface("notification_type", notificationType).
			Msg("Unable to validate notification request")

		return nil, err
	}

	userResult, err := n.user(ctx, request.Email)
	if err != nil || userResult == nil {
		n.logger.Error().
			Err(err).
			Interface("notification_type", notificationType).
			Interface("email", request.Email).
			Msg("Unable to lookup user for notification")

		return nil, err
	}

	if !userResult.IsPhoneNumberVerified || userResult.PhoneNumber == "" {
		return &notifyResponse{Error: errUserPhoneNotVerified.Error()}, nil
	}

	n.logger.Info().
		Interface("user", userResult).
		Interface("message", request).
		Str("notification_type", notificationType).
		Msg("Sending message to user")

	msg := request.Message
	alertID := ""
	incidentID := grafana.IncidentID(msg)

	// Get details
	alertGroup, err := n.grafanaSvc.IncidentDetails(ctx, incidentID)

	if err != nil {
		n.logger.Error().
			Int("incident_id", incidentID).
			Err(err).
			Msg("Failed to load incident details via Grafana API")
	}

	if alertGroup != nil {
		alertID = alertGroup.AlertID()

		n.logger.Info().
			Int("incident_id", incidentID).
			Str("alert_id", alertID).
			Msg("Loaded incident details from Grafana OnCall API")

		if notificationType == phoneCallNotify {
			// Render message
			msg = alertGroup.PhoneCall(msg)
		}

		if notificationType == smsTextNotify {
			// Render message
			msg = alertGroup.SMSText(msg)
		}
	}

	if notificationType == phoneCallNotify {
		err = n.plugin.CallPhone(ctx, *userResult, alertID, msg)
	} else {
		err = n.plugin.SendSms(ctx, *userResult, alertID, msg)
	}

	if err != nil {
		n.logger.Error().
			Err(err).
			Interface("email", request.Email).
			Interface("phone", userResult.PhoneNumber).
			Str("notification_type", notificationType).
			Int("incident_id", incidentID).
			Str("alert_id", alertID).
			Msg("Unable to send notification")

		if err == sql.ErrNoRows {
			err = errInvalidPushToken
		}

		n.logAction(userResult.ID,
			userResult.PhoneNumber, notificationType, false, err.Error())

		return nil, err
	}

	successStr := sentSuccess
	if notificationType == phoneCallNotify {
		successStr = callSuccess
	}

	n.logger.Info().
		Interface("notification_type", notificationType).
		Interface("phone", userResult.PhoneNumber).
		Msg(successStr)

	n.logAction(userResult.ID, userResult.PhoneNumber, notificationType, true, successStr)

	return &notifyResponse{}, nil
}

func (n *Notify) validate(_ context.Context, req *http.Request, _ string) (*NotifyRequest, error) {
	err := req.ParseForm()
	if err != nil {
		return nil, err
	}

	formData := req.PostForm

	n.logger.Info().
		Interface("request", formData).
		Str("url", req.URL.String()).
		Msg("Incoming request")

	email := formData.Get("email")
	if !strings.Contains(email, "@") {
		return nil, errInvalidEmail
	}

	alertText := formData.Get("message")

	if alertText == "" {
		return nil, errEmptyMessage
	}

	return &NotifyRequest{Email: email, Message: alertText}, nil
}

//nolint:interfacer
func (n *Notify) user(ctx context.Context, email string) (*user.Item, error) {
	// Lookup in DB, if not found - use user from local cache
	userResult, err := n.users.WithEmail(ctx, email)
	if err == nil {
		n.cache.Store(email, userResult)
		return userResult, nil
	}

	if cachedItem, ok := n.cache.Load(email); ok {
		return cachedItem.(*user.Item), nil
	}

	return nil, errUserNotFound
}

func (n *Notify) logAction(userID, phoneNumber, route string, success bool, msg string) {
	if n.actionsLog == nil {
		return
	}

	record := &events.Record{
		Timestamp: time.Now(),
		UserID:    userID,
		Recipient: phoneNumber,
		Channel:   route,
		Success:   success,
		Msg:       msg,
	}

	_ = n.actionsLog.Add(record)
}
