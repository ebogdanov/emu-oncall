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

	flag.String("app.auth_token", os.Getenv("AUTH_TOKEN"), "Auth token used for API request authorizations")
	flag.String("app.port", ":8880", "Application port")
	flag.String("app.hostname", "", "Application hostname")
	flag.String("app.log_level", "debug", "The minimum logging level")
	flag.String("app.log_path", "", "Path to file where logs should be stored")
	flag.Duration("app.write_timeout", defaultWriteTimeout, "The maximum duration before timing out writes of the server response")

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
		Port:         viper.GetString("app.port"),
		Version:      viper.GetString("app.version"),
		Hostname:     viper.GetString("app.hostname"),
		LogLevel:     viper.GetString("app.log_level"),
		WriteTimeout: viper.GetDuration("app.write_timeout"),
		AuthToken:    viper.GetString("app.auth_token"),
		Plugin:       viper.Get("plugin"),
	}

	return appConfig, nil
}
