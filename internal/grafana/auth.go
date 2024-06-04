package grafana

import (
	"encoding/base64"
	"fmt"
)

type engineToken struct {
	User     string
	Password string
	Token    string
}

type authCredentials interface {
	HeaderValue() string
}

func (t *engineToken) HeaderValue() string {
	if t.Token != "" {
		return t.Token
	}

	return fmt.Sprintf("Basic %s",
		base64.StdEncoding.EncodeToString([]byte(
			fmt.Sprintf("%s:%s", t.User, t.Password))))
}
