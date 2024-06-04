package v1

import (
	"encoding/json"
	"net/http"

	"github.com/ebogdanov/emu-oncall/internal/user"
)

// ErrMessage is error message object
type ErrMessage struct {
	Msg string `json:"detail"`
}

type Users struct {
	db *user.Storage
}

func NewUsers(db *user.Storage) *Users {
	return &Users{db: db}
}

func (u *Users) ServeHTTP(response http.ResponseWriter, req *http.Request) {
	result, err := u.db.Filter(req.Context(), *req)

	if err == nil {
		resp, err1 := json.Marshal(result)

		if err1 == nil {
			response.WriteHeader(http.StatusOK)
			_, _ = response.Write(resp)

			return
		}

		err = err1
	}

	errMsg := &ErrMessage{Msg: err.Error()}
	resp, _ := json.Marshal(errMsg)

	response.WriteHeader(http.StatusInternalServerError)
	_, _ = response.Write(resp)
}
