package ui

import (
	"net/http"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) ServeHTTP(response http.ResponseWriter, _ *http.Request) {
	resp := []byte("Not implemented =(")

	response.WriteHeader(http.StatusOK)
	_, _ = response.Write(resp)
}
