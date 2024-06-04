package config

import (
	"github.com/spf13/viper"
)

type Grafana struct {
	URL         string
	User        string
	Password    string
	Token       string
	ImportUsers bool

	OnCallURL   string
	OnCallToken string

	TrackSchedules       bool
	TrackIncidentDetails bool
}

func ParseGrafana() *Grafana {
	viper.SetDefault("grafana.url", "")             // Grafana Engine URL
	viper.SetDefault("grafana.auth.user", "")       // Grafana Engine Auth User
	viper.SetDefault("grafana.auth.password", "")   // Grafana Engine Auth Password
	viper.SetDefault("grafana.auth.token", "")      // Grafana Engine Auth Token (can be used either Token or user+password)
	viper.SetDefault("grafana.import_users", false) // Load users from Grafana Engine

	viper.SetDefault("grafana.oncall.url", "")                      // Grafana OnCall URL
	viper.SetDefault("grafana.oncall.token", "")                    // Grafana OnCall API Token
	viper.SetDefault("grafana.oncall.track_schedules", false)       // Notify new duty members in Grafana OnCall
	viper.SetDefault("grafana.oncall.load_incident_details", false) // Load more incident details from Grafana OnCall

	cfg := &Grafana{
		URL:         viper.GetString("grafana.url"),
		User:        viper.GetString("grafana.auth.user"),
		Password:    viper.GetString("grafana.auth.password"),
		Token:       viper.GetString("grafana.auth.token"),
		ImportUsers: viper.GetBool("grafana.import_users"),

		OnCallURL:            viper.GetString("grafana.oncall.url"),
		OnCallToken:          viper.GetString("grafana.oncall.token"),
		TrackSchedules:       viper.GetBool("grafana.oncall.track_schedules"),
		TrackIncidentDetails: viper.GetBool("grafana.oncall.load_incident_details"),
	}

	return cfg
}
