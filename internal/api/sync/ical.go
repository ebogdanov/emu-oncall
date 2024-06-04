package sync

import (
	"encoding/json"
	"github.com/ebogdanov/emu-oncall/internal/grafana"
	"net/http"
)

type ICal struct{}

func NewICal() *ICal {
	return &ICal{}
}

func (i *ICal) ServeHTTP(response http.ResponseWriter, _ *http.Request) {

	// 1. Read body
	// 2. Insert data somewhere
	result := grafana.Ok{Success: true}

	resp, err := json.Marshal(result)
	if err == nil {
		response.WriteHeader(http.StatusOK)
		_, _ = response.Write(resp)
	}
}
