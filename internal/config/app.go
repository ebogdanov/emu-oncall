package config

import (
	"flag"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	defaultWriteTimeout = 60 * time.Second
)

type App struct {
	Debug        bool
	ConfigFile   string
	Env          string
	Hostname     string
	Name         string
	LogLevel     string
	Port         string
	Version      string
	WriteTimeout time.Duration
}

func ParseFlags() (*App, error) {
	flag.Bool("debug", true, "Enable debug mode")
	flag.String("app.port", ":8080", "Application port")
	flag.String("app.hostname", "localhost", "Application hostname")
	flag.String("app.env", "dev", "Application environment")
	flag.String("app.log_level", "debug", "The minimum logging level")
	flag.String("config", "config.yaml", "Path to config file")
	flag.Duration("app.write_timeout", defaultWriteTimeout, "The maximum duration before timing out writes of the server response")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return nil, err
	}

	appConfig := &App{
		Debug:        viper.GetBool("debug"),
		Env:          viper.GetString("app.env"),
		Port:         viper.GetString("app.port"),
		ConfigFile:   viper.GetString("app.config"),
		Version:      viper.GetString("app.version"),
		Hostname:     viper.GetString("app.hostname"),
		LogLevel:     viper.GetString("app.log_level"),
		WriteTimeout: viper.GetDuration("app.write_timeout"),
	}

	viper.AddConfigPath("config") // path to look for the config file in
	viper.AddConfigPath(".")

	viper.SetConfigFile(appConfig.ConfigFile)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	return appConfig, nil
}
