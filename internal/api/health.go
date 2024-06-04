package api

import (
	"encoding/json"
	"net/http"
)

type HealthResponse struct {
	HTTPCode int    `json:"http_code"`
	Version  string `json:"version"`
	Ping     bool   `json:"ping_db"`
}

type callback func(*HealthResponse)

func NewHealthChecker(version string, check callback) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := &HealthResponse{
			HTTPCode: http.StatusInternalServerError,
			Version:  version,
			Ping:     false,
		}

		check(result)

		resp, _ := json.Marshal(result)

		w.WriteHeader(result.HTTPCode)
		_, _ = w.Write(resp)
	})
}
