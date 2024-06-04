package v1

import (
	"context"
	"net/http"
)

// ErrMessage is error message object
type ErrMessage struct {
	Msg string `json:"detail"`
}

type Handler interface {
	Request(ctx context.Context, req http.Request) (interface{}, error)
	ServeHTTP(http.ResponseWriter, *http.Request)
}
