package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	URL string `json:"url"`
}

type Info struct{}

func NewInfo() *Info {
	return &Info{}
}

func (i *Info) ServeHTTP(response http.ResponseWriter, req *http.Request) {
	result := &Response{URL: fmt.Sprintf("https://%s/", req.URL.Hostname())}

	resp, err := json.Marshal(result)
	if err == nil {
		response.WriteHeader(http.StatusOK)
		_, _ = response.Write(resp)
	}
}
