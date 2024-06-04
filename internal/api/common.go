package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ebogdanov/emu-oncall/internal/metrics"

	"github.com/ebogdanov/emu-oncall/internal/token"
	"github.com/rs/zerolog/log"
)

type errResponse struct {
	Error string `json:"error,omitempty"`
}

type responseRecorder struct {
	statusCode int
	http.ResponseWriter
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s [404] - %s %s %s", r.RemoteAddr, r.Method, r.RequestURI, r.UserAgent())

	w.WriteHeader(http.StatusNotFound)
	resp, _ := json.Marshal(&errResponse{Error: fmt.Sprintf("path %s is not found", r.RequestURI)})
	_, _ = w.Write(resp)
}

func jsonContentTypeHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

func verificationHandler(t token.Service, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := t.Verify(r)

		if err == nil {
			next.ServeHTTP(w, r)
			return
		}

		resp, _ := json.Marshal(&errResponse{Error: err.Error()})
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write(resp)
	})
}

func metricsMiddleware(promMetrics *metrics.Storage, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rr := &responseRecorder{ResponseWriter: w}
		next.ServeHTTP(rr, r)

		log.Printf("%s [%d] - %s %s %s", r.RemoteAddr, rr.statusCode, r.Method, r.RequestURI, r.UserAgent())

		status := fmt.Sprintf("%d", rr.statusCode)
		uri := strings.TrimRight(r.URL.Path, "/")

		duration := time.Since(start)

		log.Info().
			Str("component", "httpserver").
			Str("remote_addr", r.RemoteAddr).
			Int("response_code", rr.statusCode).
			Str("method", r.Method).
			Str("request_uri", r.RequestURI).
			Dur("request_time", duration).
			Str("user_agent", r.UserAgent()).
			Msg("Processed request")

		promMetrics.HTTPRequestDuration.WithLabelValues(uri, r.Method, status).Observe(duration.Seconds())
	})
}
