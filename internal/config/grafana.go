package config

import (
	"github.com/spf13/viper"
)

type Grafana struct {
	URL             string
	Token           string
	IncidentDetails bool
	OnCall          *OnCall

	Schedules map[string]*ScheduleEntry
}

type ScheduleEntry struct {
	Name        string
	Team        string
	Transport   string
	CallbackURL string
	IcalURL     string
	Details     bool
}

type NotificationTemplate struct {
	Start string
	Title string
	End   string
}

type OnCall struct {
	URL   string
	Token string
}

func ParseGrafana() *Grafana {
	viper.SetDefault("grafana.url", "")          // Grafana Engine URL
	viper.SetDefault("grafana.header_token", "") // Grafana Engine Auth Token (can be used either Token or user+password in Base64 format)

	viper.SetDefault("grafana.oncall.url", "")          // Grafana OnCall URL
	viper.SetDefault("grafana.oncall.header_token", "") // Grafana OnCall API Token
	viper.SetDefault("grafana.oncall.schedules", map[string]string{})
	viper.SetDefault("grafana.oncall.incident_details", false) // Load or not incident details from OnCall to make it detailed

	cfg := &Grafana{
		URL:   viper.GetString("grafana.url"),
		Token: viper.GetString("grafana.header_token"),

		OnCall: &OnCall{
			URL:   viper.GetString("grafana.oncall.url"),
			Token: viper.GetString("grafana.oncall.header_token"),
		},
		IncidentDetails: viper.GetBool("grafana.oncall.incident_details"),
	}

	cfg.Schedules = make(map[string]*ScheduleEntry)

	// Messy section with parsing YAML map
	configSchedules := viper.GetStringMap("grafana.oncall.schedules")
	for name, entry := range configSchedules {
		if _, ok := entry.(map[string]interface{}); !ok {
			continue
		}

		entry := entry.(map[string]interface{})

		item := &ScheduleEntry{
			Name: name,
		}

		item.Details = entry["details"].(bool)
		if _, ok := entry["details"].(bool); ok {
			item.Details = entry["details"].(bool)
		}

		if _, ok := entry["name"].(string); ok {
			item.Name = entry["name"].(string)
		}

		if _, ok := entry["team"].(string); ok {
			item.Name = entry["team"].(string)
		}

		if _, ok := entry["notify"].(string); ok {
			item.Transport = entry["notify"].(string)
		}

		if _, ok := entry["ical_url"].(string); ok {
			item.IcalURL = entry["ical_url"].(string)
		}

		if _, ok := entry["callback_url"].(string); ok {
			item.CallbackURL = entry["callback_url"].(string)
		}

		cfg.Schedules[name] = item
	}

	return cfg
}
