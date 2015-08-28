package app

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gorilla/context"
)

// Начальная авторизация
// Получаем {username:…, password:…}, отдаём token
func (app *App) AuthInit(r *http.Request) (int, interface{}) {
	reqData := &struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}
	mustBeOK(json.NewDecoder(r.Body).Decode(reqData))

	ffResp, err := http.PostForm(
		app.BackendAPIRoot+"/v1/session",
		url.Values{
			"username": {reqData.Username},
			"password": {reqData.Password},
		},
	)
	mustBeOK(err)

	defer ffResp.Body.Close()

	var m json.RawMessage
	mustBeOK(json.NewDecoder(ffResp.Body).Decode(&m))

	s := &struct {
		Users *struct {
			Username string `json:"username"`
		} `json:"users"`
		AuthToken string `json:"authToken"`
	}{}
	mustBeOK(json.Unmarshal(m, s))
	if s.Users == nil {
		return http.StatusForbidden, "Authorization failed on backend"
	}

	username, authToken := s.Users.Username, s.AuthToken

	// пролучаем наш user id
	var userID int
	err = app.DB.QueryRow("select id from users where username = $1", username).Scan(&userID)
	if mustBeOKOr(err, sql.ErrNoRows) != nil {
		// запись не найдена
		// вставляем запись, ошибки не проверяем!
		app.DB.Exec("insert into users (username, ff_token) values ($1, $2)", username, authToken)
		// и снова достаём id
		mustBeOK(app.DB.QueryRow("select id from users where username = $1", username).Scan(&userID))
	}

	return http.StatusOK, H{"token": app.LocalAuthTokens.Add(userID), "ttl": app.LocalAuthTokens.ItemTTL.Seconds()}
}

// Обновление авторизации
// Ничего не получаем, ориентируемся на заголовок, отдаём новый token
func (app *App) AuthRefresh(r *http.Request) (int, interface{}) {
	userID := context.Get(r, "UserID").(int)
	return http.StatusOK, H{"token": app.LocalAuthTokens.Add(userID), "ttl": app.LocalAuthTokens.ItemTTL.Seconds()}
}
