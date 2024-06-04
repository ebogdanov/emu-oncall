package routes

import (
	"github.com/ebogdanov/emu-oncall/internal/token"
	"net/http"
)

func notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
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

		switch err.Error() {
		case "Invalid auth token":
			http.Error(w, err.Error(), http.StatusUnauthorized)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})
}
