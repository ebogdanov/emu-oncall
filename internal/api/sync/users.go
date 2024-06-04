package sync

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/ebogdanov/emu-oncall/internal/grafana"
	"github.com/ebogdanov/emu-oncall/internal/logger"
	"github.com/ebogdanov/emu-oncall/internal/user"
)

const (
	comp = "grafana-sync-users"
)

type Users struct {
	log      *logger.Instance
	repoUser *user.Storage
}

func NewUsers(l *logger.Instance, s *user.Storage) *Users {
	return &Users{log: l, repoUser: s}
}

func (u *Users) ServeHTTP(response http.ResponseWriter, req *http.Request) {
	var list []grafana.User

	body, err := io.ReadAll(req.Body)
	if err != nil {
		u.log.Error().
			Str("component", comp).
			Err(err).
			Msg("Unable to read users body")

		u.sendError(response, err, http.StatusBadRequest)
		return
	}

	// Parse response
	err = json.Unmarshal(body, &list)
	if err != nil {
		u.log.Error().
			Str("component", comp).
			Err(err).
			Msg("Unable to parse JSON")

		u.sendError(response, err, http.StatusBadRequest)
		return
	}

	// for every Item add userItem
	for _, item := range list {
		_, err = u.repoUser.Insert(req.Context(), item.ID, item.Name, item.Login, item.Email, item.IsAdmin, !item.IsDisabled)

		if err != nil {
			u.log.Error().
				Str("component", comp).
				Interface("user", item).
				Err(err).
				Msg("unable to update user")

			u.sendError(response, err, http.StatusInternalServerError)
			return
		}
	}

	result := &grafana.Ok{Success: true}

	resp, err := json.Marshal(result)
	if err == nil {
		response.WriteHeader(http.StatusOK)
		_, _ = response.Write(resp)
	}
}

func (u *Users) sendError(response http.ResponseWriter, err error, httpCode int) {
	response.WriteHeader(httpCode)

	result := &grafana.Err{Message: err.Error()}
	resp, err1 := json.Marshal(result)

	if err1 != nil {
		u.log.Error().
			Str("component", comp).
			Err(err).
			Msg("unable to serialize error message to json")

		resp = []byte("{\"success\": false, \"msg\": \"internal error: " + err.Error() + "\"}")
	}
	_, _ = response.Write(resp)
}
