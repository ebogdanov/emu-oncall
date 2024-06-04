package api

import (
	"net/http"
	"net/http/pprof"

	"github.com/ebogdanov/emu-oncall/internal/ui"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/ebogdanov/emu-oncall/internal/metrics"

	v1 "github.com/ebogdanov/emu-oncall/internal/api/v1"

	"github.com/ebogdanov/emu-oncall/internal/token"

	"github.com/gorilla/mux"
)

type Handlers struct {
	user        *v1.Users
	info        *v1.Info
	integration *v1.Integration
	notify      *v1.Notify
	token       token.Service
	oncall      *ui.App
}

type Routes struct {
	handlers    *Handlers
	r           *mux.Router
	promMetrics *metrics.Storage
}

func NewHandlers(u *v1.Users, i *v1.Info, in *v1.Integration, n *v1.Notify, t token.Service, a *ui.App) *Handlers {
	return &Handlers{
		user:        u,
		info:        i,
		integration: in,
		notify:      n,
		token:       t,
		oncall:      a,
	}
}

func NewRoutes(h *Handlers, pm *metrics.Storage) *Routes {
	d := &Routes{
		handlers:    h,
		promMetrics: pm,
	}
	d.Create()

	return d
}

func (d *Routes) Create() *mux.Router {
	d.r = mux.NewRouter()

	d.addAPIv1()
	d.r.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	return d.r
}

func (d *Routes) addAPIv1() *Routes {
	d.r.Use(func(next http.Handler) http.Handler {
		return metricsMiddleware(d.promMetrics, next)
	})

	router := d.r.PathPrefix("/api").PathPrefix("/v1").Subrouter()
	router.Use(jsonContentTypeHandler)

	router.Use(func(next http.Handler) http.Handler {
		return verificationHandler(d.handlers.token, next)
	})

	/*
	   api/v1/users -> GET
	   api/v1/info/ -> GET
	   api/v1/make_call -> POST
	   api/v1/send_sms -> POST
	   api/v1/integrations/ -> GET+POST
	*/
	router.PathPrefix("/users").Handler(d.handlers.user).Methods(http.MethodGet)
	router.PathPrefix("/info").Handler(d.handlers.info).Methods(http.MethodGet)
	router.PathPrefix("/integrations").Handler(d.handlers.integration).Methods(http.MethodGet, http.MethodPost)

	router.PathPrefix("/make_call").Handler(d.handlers.notify).Methods(http.MethodPost)
	router.PathPrefix("/send_sms").Handler(d.handlers.notify).Methods(http.MethodPost)

	// todo /ack

	return d
}

func (d *Routes) AttachMetrics() *Routes {
	d.r.Handle("/metrics", promhttp.Handler())

	return d
}

func (d *Routes) AddHealthCheck(healthHandler http.Handler) *Routes {
	if healthHandler != nil {
		d.r.Handle("/health", healthHandler)
	}

	return d
}

func (d *Routes) AttachProfiler(debugMode bool) *Routes {
	if !debugMode {
		return d
	}

	d.r.HandleFunc("/debug/pprof/", pprof.Index)

	d.r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	d.r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	d.r.HandleFunc("/debug/pprof/trace", pprof.Trace)

	d.r.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	d.r.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	d.r.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	d.r.Handle("/debug/pprof/block", pprof.Handler("block"))
	d.r.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
	d.r.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))

	return d
}

func (d *Routes) AttachOnCallApp() *Routes {
	d.r.Handle("/a/grafana-oncall-ui/", d.handlers.oncall)

	return d
}

func (d *Routes) R() *mux.Router {
	return d.r
}
