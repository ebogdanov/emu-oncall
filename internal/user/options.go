package user

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	strTrue = "TRUE"
)

const (
	roleAdmin    = "admin"
	roleUser     = "user"
	roleObserver = "observer"
)

var (
	rolesMap = map[string]string{
		"0": roleAdmin,
		"1": roleUser,
		"2": roleObserver}
)

type Options struct {
	UserID   string
	Username string
	Page     int
	Email    string
	Short    bool
	Roles    []string
	Limit    int
}

func (s *Storage) fromHTTPRequest(req http.Request) *Options {
	params := &Options{}

	// Process /api/v1/users/USERID requests
	parts := strings.Split(req.URL.EscapedPath(), "/")

	l := len(parts)
	if l > 0 && parts[l-1] != "users" {
		params.UserID = parts[l-1]

		return params
	}

	query := req.URL.Query()

	if query.Has("username") {
		params.UserID = query.Get("username")
	}

	if query.Has("page") {
		pageNumber, err := strconv.Atoi(query.Get("page"))
		if err != nil {
			pageNumber = 1
		}
		params.Page = pageNumber
	}

	if query.Has("short") {
		params.Short = false
		if strings.ToUpper(query.Get("short")) == strTrue {
			params.Short = true
		}
	}

	if query.Has("email") {
		params.Email = query.Get("email")
	}

	if query.Has("roles") && len(query["roles"]) > 0 {
		params.Roles = []string{}

		for _, roleID := range query["roles"] {
			if roleName, ok := rolesMap[roleID]; ok {
				params.Roles = append(params.Roles, roleName)
			}
		}
	}

	return params
}
