package logger

import (
	"fmt"
	"os"

	"github.com/ebogdanov/emu-oncall/internal/config"
	"github.com/rs/zerolog"
)

func Init(cfg *config.App) zerolog.Logger {
	level, errD := zerolog.ParseLevel(cfg.LogLevel)
	if errD != nil {
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)

	var log zerolog.Logger

	if cfg.LogPath != "" {
		log = initToFile(cfg.LogPath)
	} else {
		log = initStdOut()
	}

	startMsg := log.Info()

	if cfg.Hostname != "" {
		startMsg = startMsg.Str("hostname", cfg.Hostname)
	}
	startMsg = startMsg.Str("version", cfg.Version)

	if errD != nil {
		startMsg = startMsg.Str("warn", fmt.Sprintf("%s is incorrect log level", cfg.LogLevel))
	} else {
		startMsg = startMsg.Str("log_level", level.String())
	}
	startMsg.Msg("Starting application")

	return log
}

func initStdOut() zerolog.Logger {
	return zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func initToFile(filePath string) zerolog.Logger {
	_, err := os.OpenFile(
		filePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	if err != nil {
		panic(err)
	}

	return zerolog.New(os.Stdout).With().Timestamp().Logger()
}
