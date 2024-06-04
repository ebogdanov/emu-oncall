package v1

import (
	"encoding/json"
	"github.com/ebogdanov/emu-oncall/internal/user"
	"net/http"
)

type Users struct {
	db *user.Storage
}

func NewUsers(db *user.Storage) *Users {
	return &Users{db: db}
}

func (u *Users) ServeHTTP(response http.ResponseWriter, req *http.Request) {
	result, err := u.db.ByFilter(req.Context(), *req)

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
