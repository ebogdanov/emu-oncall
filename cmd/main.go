package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ebogdanov/emu-oncall/internal/ui"

	"github.com/ebogdanov/emu-oncall/internal/api"

	"github.com/ebogdanov/emu-oncall/internal/db"
	"github.com/ebogdanov/emu-oncall/internal/events"
	"github.com/ebogdanov/emu-oncall/internal/grafana"
	"github.com/ebogdanov/emu-oncall/internal/logger"
	"github.com/ebogdanov/emu-oncall/internal/metrics"
	"github.com/ebogdanov/emu-oncall/internal/token"
	"github.com/ebogdanov/emu-oncall/internal/user"

	v1 "github.com/ebogdanov/emu-oncall/internal/api/v1"
	"github.com/ebogdanov/emu-oncall/internal/config"
	"github.com/ebogdanov/emu-oncall/plugin"

	"github.com/ebogdanov/emu-oncall/internal/graceful"
)

var (
	appConfig *config.App

	err error
)

func main() {
	appConfig, err = config.Parse()
	if err != nil {
		fmt.Printf("fatal error: %v", err)

		os.Exit(10)
	}

	graceful.Init()

	log := logger.Init(appConfig)

	dbCfg := config.ParseDB()
	grafanaCfg := config.ParseGrafana()

	promMetrics := metrics.NewMetrics()

	sqlShard, err := db.Init(dbCfg, promMetrics)
	if err != nil {
		log.Fatal().Err(err).
			Str("addr", dbCfg.Addr).
			Int("port", dbCfg.Port).
			Str("user", dbCfg.User).
			Msg("unable connect to database")
	}

	// Now you can use for testing:
	notifier := plugin.NewTextOutput(log)

	userStorage := user.NewStorage(sqlShard, log)

	// Services
	actions := events.New(sqlShard, log)
	grafanaConnect := grafana.New(grafanaCfg, notifier, userStorage, log, promMetrics)

	healthCheck := func(result *api.HealthResponse) {
		result.Ping = sqlShard.Ping() == nil

		if result.Ping {
			result.HTTPCode = http.StatusOK
		}
	}

	info := v1.NewInfo(appConfig)
	usersV1 := v1.NewUsers(userStorage)
	tokenSrv := token.NewFromConfig(appConfig)
	notifyV1 := v1.NewNotify(userStorage, notifier, log, actions, grafanaConnect, promMetrics)
	integrationV1 := v1.NewIntegration(appConfig.Hostname, promMetrics)
	onCallUI := ui.New(appConfig, log)

	handlers := api.NewHandlers(usersV1, info, integrationV1, notifyV1, tokenSrv, onCallUI)
	health := api.NewHealthChecker(appConfig.Version, healthCheck)

	routes := api.NewRoutes(handlers, promMetrics)
	routes.AttachMetrics().
		AddHealthCheck(health).
		AttachProfiler(appConfig.Debug).
		AttachOnCallApp()

	server := &http.Server{
		Addr:         appConfig.Port,
		Handler:      routes.R(),
		ReadTimeout:  appConfig.WriteTimeout,
		WriteTimeout: appConfig.WriteTimeout + 100*time.Microsecond,
	}

	ctx := context.Background()
	go func() {
		actions.Listen(ctx)
	}()

	graceful.AddCallback(func() error {
		ctx.Done()
		actions.Stop()

		log.Info().Msg("server was gracefully stopped")

		return nil
	})

	go func() {
		grafanaConnect.Start(ctx)
	}()

	go func() {
		log.Info().Msgf("HTTP Server: listening on %s", appConfig.Port)

		err = server.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("failed to bind server")
		}
	}()

	err = graceful.WaitShutdown()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to gracefully shutdown the server")
	}
}
