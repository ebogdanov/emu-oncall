package token

import (
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"
)

type Service interface {
	Verify(r *http.Request) error
}

type envTokenService struct {
	authToken string
}

func NewEnv() Service {
	return &envTokenService{authToken: os.Getenv("AUTH_TOKEN")}
}

func (s *envTokenService) Verify(r *http.Request) error {
	if s.authToken == "" {
		return nil
	}

	authToken := r.Header.Get("Authorization")

	authToken = strings.TrimPrefix(authToken, "Bearer")

	if authToken == s.authToken {
		return nil
	}

	return errors.New("invalid auth token")
}
