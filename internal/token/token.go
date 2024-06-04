package token

import (
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/ebogdanov/emu-oncall/internal/config"
)

var (
	errInvalidAuthToken = errors.New("invalid auth token")
)

type Service interface {
	Verify(r *http.Request) error
}

type envTokenService struct {
	authToken string
}

func NewFromConfig(app *config.App) Service {
	return &envTokenService{authToken: app.AuthToken}
}

func (s *envTokenService) Verify(r *http.Request) error {
	if s.authToken == "" {
		return nil
	}

	authToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

	if authToken != s.authToken {
		return errInvalidAuthToken
	}

	return nil
}
