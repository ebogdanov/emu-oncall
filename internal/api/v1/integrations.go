package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	webhookURL = "/v1/formatted_webhook/oNDZUWVBQa0x5SFNmSmhGND0K/"
)

type Integration struct {
	hostname string
}

type IntegrationsResponse struct {
	Count    int                 `json:"count"`
	Next     interface{}         `json:"next"`
	Previous interface{}         `json:"previous"`
	Results  []IntegrationResult `json:"results"`
}

type IntegrationResult struct {
	MaintenanceMode      interface{}   `json:"maintenance_mode"`
	MaintenanceStartedAt interface{}   `json:"maintenance_started_at"`
	MaintenanceEndAt     interface{}   `json:"maintenance_end_at"`
	ID                   string        `json:"id"`
	Name                 string        `json:"name"`
	TeamID               interface{}   `json:"team_id"`
	Link                 string        `json:"link"`
	Type                 string        `json:"type"`
	Templates            *Templates    `json:"templates"`
	Heartbeat            *Heartbeat    `json:"heartbeat"`
	DefaultRoute         *DefaultRoute `json:"default_route"`
}

type ChannelEnabled struct {
	ID      interface{} `json:"id"`
	Enabled bool        `json:"enabled"`
}

type SlackChannelEnabled struct {
	ChannelID interface{} `json:"channel_id"`
	Enabled   bool        `json:"enabled"`
}

type DefaultRoute struct {
	ID                string               `json:"id"`
	EscalationChainID interface{}          `json:"escalation_chain_id"`
	Slack             *SlackChannelEnabled `json:"slack"`
	Telegram          *ChannelEnabled      `json:"telegram"`
	Email             *ChannelEnabled      `json:"email"`
	MsTeams           *ChannelEnabled      `json:"msteams"`
}

type Templates struct {
	GroupingKey       interface{}        `json:"grouping_key"`
	ResolveSignal     interface{}        `json:"resolve_signal"`
	AcknowledgeSignal interface{}        `json:"acknowledge_signal"`
	Slack             *MessengerTemplate `json:"slack"`
	Web               *MessengerTemplate `json:"web"`
	Sms               *TitleTemplate     `json:"sms"`
	PhoneCall         *TitleTemplate     `json:"phone_call"`
	Telegram          *MessengerTemplate `json:"telegram"`
	Email             *EmailTemplate     `json:"email"`
	MsTeams           *MessengerTemplate `json:"msteams"`
}

type TitleTemplate struct {
	Title interface{} `json:"title"`
}

type EmailTemplate struct {
	*TitleTemplate
	Message interface{} `json:"message"`
}

type MessengerTemplate struct {
	*EmailTemplate
	ImageURL interface{} `json:"image_url"`
}

type Heartbeat struct {
	Link string `json:"link"`
}

func NewIntegration(host string) *Integration {
	return &Integration{
		hostname: fmt.Sprintf("https://%s/oncall/integrations", host),
	}
}

func (i *Integration) ServeHTTP(response http.ResponseWriter, req *http.Request) {
	// This endpoint is not used, so that IDs are hardcoded
	item := &IntegrationResult{
		ID:   "F566XENITUQK4",
		Name: fmt.Sprintf("OnCall Cloud Heartbeat %s", i.hostname),
		Link: i.hostname + webhookURL,
		Type: "formatted_webhook",
		Templates: &Templates{
			Sms:       &TitleTemplate{},
			Web:       &MessengerTemplate{},
			Email:     &EmailTemplate{},
			Slack:     &MessengerTemplate{},
			MsTeams:   &MessengerTemplate{},
			Telegram:  &MessengerTemplate{},
			PhoneCall: &TitleTemplate{},
		},
		Heartbeat: &Heartbeat{Link: i.hostname + webhookURL + "heartbeat/"},
		DefaultRoute: &DefaultRoute{
			ID:                "U7SF1KOR2KU6C",
			Slack:             &SlackChannelEnabled{},
			Email:             &ChannelEnabled{},
			MsTeams:           &ChannelEnabled{},
			Telegram:          &ChannelEnabled{},
			EscalationChainID: nil,
		},
	}

	result := &IntegrationsResponse{
		Count:    1,
		Next:     nil,
		Previous: nil,
		Results:  []IntegrationResult{*item},
	}

	if req.Method == http.MethodPost {
		response.WriteHeader(http.StatusCreated)
	} else {
		response.WriteHeader(http.StatusOK)
	}

	resp, err := json.Marshal(result)
	if err == nil {
		_, _ = response.Write(resp)
	}
}
