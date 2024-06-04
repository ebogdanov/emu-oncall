package logger

import (
	"os"

	"github.com/ebogdanov/emu-oncall/internal/config"
	"github.com/rs/zerolog"
)

type Instance struct {
	zerolog.Logger
}

func Init(cfg *config.App) *Instance {
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)

	instance := zerolog.New(os.Stdout)

	instance.Info().
		Str("host", cfg.Hostname).
		Str("version", cfg.Version).
		Str("env", cfg.Env).
		Msgf("Starting %s", cfg.Name)

	return &Instance{instance}
}
