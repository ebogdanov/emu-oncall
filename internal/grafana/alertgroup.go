package grafana

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	regexpInstance = regexp.MustCompile(`\*instance\*:\s+(.*)\s`)
)

type AlertGroup struct {
	ID                       string              `json:"pk"`
	AlertsCount              int                 `json:"alerts_count"`
	InsideOrganizationNumber int                 `json:"inside_organization_number"`
	AlertReceiveChannel      alertReceiveChannel `json:"alert_receive_channel"`
	Resolved                 bool                `json:"resolved"`
	ResolvedBy               int                 `json:"resolved_by"`
	ResolvedByUser           string              `json:"resolved_by_user"`
	ResolvedAt               time.Time           `json:"resolved_at"`
	AcknowledgedAt           time.Time           `json:"acknowledged_at"`
	Acknowledged             bool                `json:"acknowledged"`
	AcknowledgedOnSource     bool                `json:"acknowledged_on_source"`
	AcknowledgedByUser       acknowledgedByUser  `json:"acknowledged_by_user"`
	Silenced                 bool                `json:"silenced"`
	SilencedByUser           string              `json:"silenced_by_user"`
	SilencedAt               time.Time           `json:"silenced_at"`
	SilencedUntil            time.Time           `json:"silenced_until"`
	StartedAt                time.Time           `json:"started_at"`
	RelatedUsers             []relatedUser       `json:"related_users"`
	RenderForWeb             renderWeb           `json:"render_for_web"`
	RenderForClassicMarkdown renderMarkdown      `json:"render_for_classic_markdown"`
	DependentAlertGroups     []string            `json:"dependent_alert_groups"`
	RootAlertGroup           string              `json:"root_alert_group"`
	Status                   int                 `json:"status"`
}

type Render interface {
	AlertID() string
	PhoneCall(string) string
	SMSText(string) string
	Raw() interface{}
}

func (a *AlertGroup) AlertID() string {
	return a.ID
}

func (a *AlertGroup) PhoneCall(originalMessage string) string {
	matches := regexpInstance.FindAllStringSubmatch(a.RenderForClassicMarkdown.Message, -1)
	if len(matches) > 0 {
		originalMessage = strings.Replace(originalMessage, ", alert channel", " "+matches[0][1]+", alert channel", -1)
	}

	return originalMessage
}

func (a *AlertGroup) SMSText(originalMessage string) string {
	matches := regexpInstance.FindAllStringSubmatch(a.RenderForClassicMarkdown.Message, -1)
	if len(matches) > 0 {
		return strings.Replace(originalMessage, ", alert channel", " "+matches[0][1]+", alert channel", -1)
	}

	return originalMessage
}

func (a *AlertGroup) Raw() interface{} {
	return a
}

func IncidentID(msg string) int {
	var incidentID int
	matches := regexpIncidentID.FindAllStringSubmatch(msg, -1)

	if len(matches) > 0 {
		incidentID, _ = strconv.Atoi(matches[0][1])
	}

	return incidentID
}
