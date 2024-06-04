package config

import (
	"flag"
	"os"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	defaultWriteTimeout = 60 * time.Second
)

type App struct {
	Port         string
	Debug        bool
	Hostname     string
	LogLevel     string
	Version      string
	AuthToken    string
	LogPath      string
	WriteTimeout time.Duration
	Plugin       interface{}
}

func Parse() (*App, error) {
	flag.Bool("debug", true, "Enable debug mode")
	flag.String("config", "config/config.yml", "Path to config file")

	flag.String("ui.auth_token", os.Getenv("AUTH_TOKEN"), "Auth token used for API request authorizations")
	flag.String("ui.port", ":8880", "Application port")
	flag.String("ui.hostname", "", "Application hostname")
	flag.String("ui.log_level", "debug", "The minimum logging level")
	flag.String("ui.log_path", "", "Path to file where logs should be stored")
	flag.Duration("ui.write_timeout", defaultWriteTimeout, "The maximum duration before timing out writes of the server response")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return nil, err
	}

	configFile := viper.GetString("config")

	// path to look for the config file in
	viper.AddConfigPath(".")
	viper.SetConfigFile(configFile)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	appConfig := &App{
		Debug:        viper.GetBool("debug"),
		Port:         viper.GetString("ui.port"),
		Version:      viper.GetString("ui.version"),
		Hostname:     viper.GetString("ui.hostname"),
		LogLevel:     viper.GetString("ui.log_level"),
		WriteTimeout: viper.GetDuration("ui.write_timeout"),
		AuthToken:    viper.GetString("ui.auth_token"),
		Plugin:       viper.Get("plugin"),
	}

	return appConfig, nil
}
