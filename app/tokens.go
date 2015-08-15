package app

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/davidmz/aproxy33/sqlt"
	"github.com/gorilla/context"
	"github.com/sqs/mux"
)

type TokenInfo struct {
	ID    int               `json:"id"`
	Date  time.Time         `json:"date"`
	Perms map[string]string `json:"perms"`
	App   struct {
		Title       string           `json:"title"`
		Description string           `json:"description"`
		Owner       string           `json:"owner"`
		Domains     sqlt.StringSlice `json:"domains"`
	} `json:"app"`
}

// Список активных токенов  данного юзера
func (a *App) TokensList(r *http.Request) (int, interface{}) {
	userID := context.Get(r, "UserID").(int)

	rows := mustBeOKVal(a.DB.Query(
		`select 
			t.id, t.date, t.perms, a.title, a.description, a.domains, u.username 
		from 
			`+a.DBTablePrefix+`atokens t 
			join `+a.DBTablePrefix+`apps a on a.id = t.app_id
			join `+a.DBTablePrefix+`users a on u.id = t.user_id
		where t.user_id = $1 order by t.date desc`,
		userID,
	)).(*sql.Rows)

	defer rows.Close()
	list := []TokenInfo{}
	for rows.Next() {
		ti := TokenInfo{Perms: make(map[string]string)}
		var pNames sqlt.StringSlice
		mustBeOK(rows.Scan(&ti.ID, &ti.Date, &pNames, &ti.App.Title, &ti.App.Description, &ti.App.Domains, &ti.App.Owner))
		for _, p := range pNames {
			ti.Perms[p] = a.BackendAPIMethods[p]
		}
		list = append(list, ti)
	}
	mustBeOK(rows.Err())

	return http.StatusOK, list
}

// Удаление токена
func (a *App) TokensDelete(r *http.Request) (int, interface{}) {
	userID := context.Get(r, "UserID").(int)
	id := mustBeOKVal(strconv.Atoi(mux.Vars(r)["id"])).(int)

	err := a.DB.QueryRow("delete from "+a.DBTablePrefix+"atokens where id = $1 and user_id = $2 returning id", id, userID).Scan()
	if mustBeOKOr(err, sql.ErrNoRows) != nil {
		return http.StatusNotFound, "Token not found"
	}

	return http.StatusOK, nil
}
