package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ebogdanov/emu-oncall/internal/db"
	"github.com/ebogdanov/emu-oncall/internal/events"
	"github.com/ebogdanov/emu-oncall/internal/grafana"
	"github.com/ebogdanov/emu-oncall/internal/logger"
	"github.com/ebogdanov/emu-oncall/internal/token"
	"github.com/ebogdanov/emu-oncall/internal/user"

	"github.com/ebogdanov/emu-oncall/internal/api/routes"
	v1 "github.com/ebogdanov/emu-oncall/internal/api/v1"
	"github.com/ebogdanov/emu-oncall/internal/config"
	"github.com/ebogdanov/emu-oncall/plugin"

	apihttp "github.com/ebogdanov/emu-oncall/internal/api/http"

	"github.com/ebogdanov/emu-oncall/internal/graceful"
)

func main() {
	appConfig, err := config.ParseFlags()
	if err != nil {
		fmt.Println(fmt.Errorf("fatal error: %v", err))

		os.Exit(10)
	}

	log := logger.Init(appConfig)

	dbCfg := config.ParseDB()
	grafanaCfg := config.ParseGrafana()

	// Or you can use for testing:
	notifier := plugin.NewTextOutput(log)

	sqlShard, err := db.Init(dbCfg)
	if err != nil {
		log.Fatal().Err(err).
			Str("addr", dbCfg.Addr).
			Int("port", dbCfg.Port).
			Str("user", dbCfg.User).
			Msg("unable connect to database")
	}

	// Data repositories
	repoUser := user.NewStorage(sqlShard, log)

	// Services
	actions := events.New(sqlShard, *log)
	grafanaConnect := grafana.New(grafanaCfg, notifier, repoUser, log)

	infoHandler := v1.NewInfo()
	userHandler := v1.NewUsers(repoUser)
	integrationHandler := v1.NewIntegration(appConfig.Hostname)
	notifyHandler := v1.NewNotify(repoUser, notifier, log, actions, grafanaConnect)

	tokenSrv := token.NewEnv()

	routeList := routes.NewRoutes(userHandler, infoHandler, integrationHandler, notifyHandler, tokenSrv)
	router := routeList.Init()

	if appConfig.Debug {
		apihttp.AttachProfiler(router)
	}

	server := &http.Server{
		Addr:         appConfig.Port,
		Handler:      router,
		ReadTimeout:  appConfig.WriteTimeout,
		WriteTimeout: appConfig.WriteTimeout + 100*time.Microsecond,
	}

	ctx := context.Background()
	graceful.AddCallback(func() error {
		return server.Shutdown(ctx)
	})

	go func() {
		log.Info().Msgf("HTTP server: listening on %s", appConfig.Port)

		err = server.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("failed to listen server")
		}
	}()
	actions.Start(ctx)
	grafanaConnect.Start(ctx)

	err = graceful.WaitShutdown()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to gracefully shutdown server")
	}

	actions.Stop()
	log.Info().Msg("server gracefully stopped")
}
