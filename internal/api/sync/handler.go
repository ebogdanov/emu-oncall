package sync

import (
	"context"
	"net/http"
)

// Err is error message object
type Err struct {
	Msg string `json:"detail"`
}

type Handler interface {
	Request(ctx context.Context, req http.Request) (interface{}, error)
	ServeHTTP(http.ResponseWriter, *http.Request)
}
