package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ebogdanov/emu-oncall/internal/config"
)

type Response struct {
	URL string `json:"url"`
}

type Info struct {
	cfg *config.App
}

func NewInfo(cfgApp *config.App) *Info {
	return &Info{cfg: cfgApp}
}

func (i *Info) ServeHTTP(response http.ResponseWriter, req *http.Request) {
	hostName := i.cfg.Hostname
	if hostName == "" {
		hostName = req.Host
	}

	if !strings.HasPrefix(hostName, "http:") && !strings.HasPrefix(hostName, "https:") {
		hostName = fmt.Sprintf("http://%s", hostName)
	}

	// OnCall do not like to handle double slashes path like 127.0.0.1//api, so that filter it if any
	result := &Response{URL: strings.TrimRight(hostName, "/")}

	resp, _ := json.Marshal(result)
	response.WriteHeader(http.StatusOK)
	_, _ = response.Write(resp)
}
