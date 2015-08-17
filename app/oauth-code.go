package app

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/davidmz/aproxy33/sqlt"
	"github.com/gorilla/context"
)

type OAuthCodeInfo struct {
	AppID       int
	UserID      int
	RedirectURI string
	Perms       []string
}

// Выдать код авторизации
// принимает данные в формате: {"app_key": "…", "redirect_uri": "…", "perms": ["…", …]}
func (a *App) OAuthNewCode(r *http.Request) (int, interface{}) {
	userID := context.Get(r, "UserID").(int)

	reqData := &struct {
		AppKey      string   `json:"app_key"`
		RedirectURI string   `json:"redirect_uri"`
		Perms       []string `json:"perms"`
	}{}
	mustBeOK(json.NewDecoder(r.Body).Decode(reqData))

	ci := &OAuthCodeInfo{
		UserID:      userID,
		RedirectURI: reqData.RedirectURI,
	}

	var domains sqlt.StringSlice

	// есть ли такое приложение?
	err := a.DB.QueryRow("select id, domains from apps where key = $1", reqData.AppKey).Scan(&ci.AppID, &domains)
	if mustBeOKOr(err, sql.ErrNoRows) != nil {
		return http.StatusNotFound, "App not found"
	}

	{
		// проверка RedirectURI
		u, err := url.Parse(ci.RedirectURI)
		if err != nil || !u.IsAbs() || u.Scheme != "http" || u.Scheme != "https" {
			return http.StatusBadRequest, "Invalid 'redirect_uri' format (must be absolute http/https URL)"
		}

		domFound := false
		for _, d := range domains {
			if d == u.Host {
				domFound = true
				break
			}
		}
		if !domFound {
			return http.StatusBadRequest, "'redirect_uri' have incorrect domain"
		}
	}

	{
		// проверка запрошенных разрешений, оставляем только разрешённые и уникальные
		u, e := map[string]struct{}{}, struct{}{}
		for _, s := range reqData.Perms {
			if _, seen := u[s]; !seen {
				if _, allowed := a.BackendAPIMethods[s]; allowed {
					reqData.Perms[len(u)] = s
					u[s] = e
				}
			}
		}
		reqData.Perms = reqData.Perms[:len(u)]
	}

	ci.Perms = reqData.Perms

	return http.StatusOK, a.OAuthCodes.Add(ci)
}
