package routes

import (
	"net/http"

	"github.com/ebogdanov/emu-oncall/internal/token"

	v1 "github.com/ebogdanov/emu-oncall/internal/api/v1"
	"github.com/gorilla/mux"
)

type Routes struct {
	// healthCh           *v1.Health
	userHandler        *v1.Users
	infoHandler        *v1.Info
	integrationHandler *v1.Integration
	notifyHandler      *v1.Notify
	tokenService       token.Service
}

func NewRoutes(u *v1.Users, i *v1.Info, in *v1.Integration, n *v1.Notify, t token.Service) *Routes {
	return &Routes{
		userHandler:        u,
		infoHandler:        i,
		integrationHandler: in,
		notifyHandler:      n,
		tokenService:       t,
	}
}

func (d *Routes) Init() *mux.Router {
	router := mux.NewRouter()

	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	d.addAPIV1Routes(router)
	// d.addHealthCheckRoute(router)

	return router
}

func (d *Routes) addAPIV1Routes(r *mux.Router) {
	router := r.PathPrefix("/grafana/api/").PathPrefix("/v1/").Subrouter()
	router.Use(jsonContentTypeHandler)

	router.Use(func(next http.Handler) http.Handler {
		return verificationHandler(d.tokenService, next)
	})

	/*
	   api/v1/users -> GET
	   api/v1/info/ -> GET
	   api/v1/make_call -> POST
	   api/v1/send_sms -> POST
	   api/v1/integrations/  -> GET+POST
	*/
	router.PathPrefix("/users").Handler(d.userHandler).Methods(http.MethodGet)
	router.PathPrefix("/info").Handler(d.infoHandler).Methods(http.MethodGet)
	router.PathPrefix("/integrations").Handler(d.integrationHandler).Methods(http.MethodGet, http.MethodPost)

	router.PathPrefix("/make_call").Handler(d.notifyHandler).Methods(http.MethodPost)
	router.PathPrefix("/send_sms").Handler(d.notifyHandler).Methods(http.MethodPost)
}

/*
func (d *Routes) addHealthCheckRoute(r *mux.Router) {
	if d.healthCh == nil {
		return
	}

	r.HandleFunc("/health", health.NewHandler(d.healthCh, "health_router"))
}

func (d *Routes) addMetricsRoute(r *mux.Router) {
	r.Handle("/metrics", promhttp.Handler())
}
*/
