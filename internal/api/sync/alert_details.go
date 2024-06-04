package sync

import (
	"encoding/json"
	"github.com/ebogdanov/emu-oncall/internal/grafana"
	"net/http"
)

type AlertDetails struct{}

func NewAlertDetails() *AlertDetails {
	return &AlertDetails{}
}

func (a *AlertDetails) ServeHTTP(response http.ResponseWriter, _ *http.Request) {

	// 1. Read body
	// 2. Insert based on config either to cache or local map (with thread safe)
	result := grafana.Ok{Success: true}

	resp, err := json.Marshal(result)
	if err == nil {
		response.WriteHeader(http.StatusOK)
		_, _ = response.Write(resp)
	}
}
