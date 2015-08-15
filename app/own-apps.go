package app

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/davidmz/aproxy33/sqlt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

type OwnApp struct {
	Key         string           `json:"key"`
	Secret      string           `json:"secret,omitempty"`
	Date        time.Time        `json:"date,omitempty"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Domains     sqlt.StringSlice `json:"domains"`
}

// Список собственных приложений данного юзера
func (a *App) AppsList(r *http.Request) (int, interface{}) {
	userID := context.Get(r, "UserID").(int)

	rows := mustBeOKVal(a.DB.Query(
		"select key, secret, date, title, description, domains from "+a.DBTablePrefix+"apps where user_id = $1 order by date desc",
		userID,
	)).(*sql.Rows)

	defer rows.Close()
	list := []OwnApp{}
	for rows.Next() {
		oa := OwnApp{}
		mustBeOK(rows.Scan(&oa.Key, &oa.Secret, &oa.Date, &oa.Title, &oa.Description, &oa.Domains))
		list = append(list, oa)
	}
	mustBeOK(rows.Err())

	return http.StatusOK, list
}

// Создание нового приложения
func (a *App) AppsNew(r *http.Request) (int, interface{}) {
	userID := context.Get(r, "UserID").(int)

	oa := &OwnApp{}
	mustBeOK(json.NewDecoder(r.Body).Decode(oa))

	oa.Title = strings.TrimSpace(oa.Title)
	oa.Description = strings.TrimSpace(oa.Description)
	domains := oa.Domains
	oa.Domains = []string{}
	for _, d := range domains {
		d = strings.TrimSpace(d)
		if d != "" {
			oa.Domains = append(oa.Domains, d)
		}
	}

	if oa.Title == "" || oa.Description == "" || len(oa.Domains) == 0 {
		return http.StatusBadRequest, "Required fields: title, description, domains"
	}

	b := make([]byte, 40)
	mustBeOKVal(rand.Read(b))

	oa.Key = fmt.Sprintf("%x", b[:10])
	oa.Secret = fmt.Sprintf("%x", b[30:])

	mustBeOKVal(a.DB.Query(
		"insert into "+a.DBTablePrefix+"apps (key, secret, title, description, domains, user_id) values ($1, $2, $3, $4, $5, $6)",
		oa.Key, oa.Secret, oa.Title, oa.Description, oa.Domains, userID,
	))

	return http.StatusOK, H{"key": oa.Key, "secret": oa.Secret}
}

// Изменение приложения
func (a *App) AppsUpdate(r *http.Request) (int, interface{}) {
	userID := context.Get(r, "UserID").(int)
	key := mux.Vars(r)["key"]

	appID := 0
	err := a.DB.QueryRow("select id from "+a.DBTablePrefix+"apps where key = $1 and user_id = $2", key, userID).Scan(&appID)
	if mustBeOKOr(err, sql.ErrNoRows) != nil {
		return http.StatusNotFound, "App not found"
	}

	oa := &OwnApp{}
	mustBeOK(json.NewDecoder(r.Body).Decode(oa))

	oa.Title = strings.TrimSpace(oa.Title)
	oa.Description = strings.TrimSpace(oa.Description)
	domains := oa.Domains
	oa.Domains = []string{}
	for _, d := range domains {
		d = strings.TrimSpace(d)
		if d != "" {
			oa.Domains = append(oa.Domains, d)
		}
	}

	if oa.Title == "" || oa.Description == "" || len(oa.Domains) == 0 {
		return http.StatusBadRequest, "Required fields: title, description, domains"
	}

	mustBeOKVal(a.DB.Query(
		"update "+a.DBTablePrefix+"apps set title = $1, description = $2, domains = $3 where id = $4",
		oa.Title, oa.Description, oa.Domains, appID,
	))

	return http.StatusOK, nil
}

// Удаление приложения
func (a *App) AppsDelete(r *http.Request) (int, interface{}) {
	userID := context.Get(r, "UserID").(int)
	key := mux.Vars(r)["key"]

	err := a.DB.QueryRow("delete from "+a.DBTablePrefix+"apps where key = $1 and user_id = $2 returning id", key, userID).Scan()
	if mustBeOKOr(err, sql.ErrNoRows) != nil {
		return http.StatusNotFound, "App not found"
	}

	return http.StatusOK, nil
}

// Публичная информация о приложении, доступна без авторизации
func (a *App) AppsInfo(r *http.Request) (int, interface{}) {
	key := mux.Vars(r)["key"]
	oa := &OwnApp{}
	err := a.DB.
		QueryRow("select key, date, title, description, domains from "+a.DBTablePrefix+"apps where key = $1", key).
		Scan(&oa.Key, &oa.Date, &oa.Title, &oa.Description, &oa.Domains)

	if mustBeOKOr(err, sql.ErrNoRows) != nil {
		return http.StatusNotFound, "App not found"
	}

	return http.StatusOK, H{"app": oa, "api": a.BackendAPIMethods}
}
