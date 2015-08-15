package app

import (
	"database/sql"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var (
	HeadersFromBackend = []string{"Content-Type", "Etag", "Last-Modified"}
	HeadersFromClient  = []string{"Content-Type", "If-None-Match", "If-Modified-Since"}
)

func (a *App) ProxyBackend(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	apiPath := strings.TrimPrefix(r.URL.Path, "/frf")
	apiMethod := r.Method + " " + apiPath

	frfToken := ""
	perms := a.BackendAPIMethods

	// Пришли ли с авторизацией?
	if _, ok := r.Header["Authorization"]; ok {
		auh := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(auh) != 2 || auh[0] != "Bearer" {
			oAuthError(H{"error": "invalid_request", "error_description": "Invalid authorization header"}, w)
			return
		}

		err := a.DB.QueryRow(
			`select t.perms, u.ff_token from 
				`+a.DBTablePrefix+`atokens t
				join `+a.DBTablePrefix+`users u on u.id = t.user_id
			where t.token = $1`,
			auh[1],
		).Scan(&perms, &frfToken)

		if err == sql.ErrNoRows {
			oAuthError(H{"error": "invalid_token", "error_description": "Invalid authorization token"}, w)
			return
		} else if err != nil {
			a.Log.WithField("error", err).Error("Database error")
			oAuthFail(H{"error": "Internal error"}, w)
			return
		}
	}

	found := false
	for m, _ := range perms {
		m = "^" + strings.Replace(regexp.QuoteMeta(m), "\\*", "[^/]+", -1) + "$"
		re := regexp.MustCompile(m)
		if re.MatchString(apiMethod) {
			found = true
			break
		}
	}

	if !found {
		oAuthError(H{"error": "invalid_request", "error_description": "API method not allowed"}, w)
		return
	}

	req, err := http.NewRequest(
		r.Method,
		a.BackendAPIRoot+strings.TrimPrefix(r.URL.RequestURI(), "/frf"),
		r.Body,
	)
	if err != nil {
		a.Log.WithField("error", err).Error("Can not create proxy request")
		oAuthFail(H{"error": "Can not create proxy request"}, w)
		return
	}

	for _, h := range HeadersFromClient {
		if hh, ok := r.Header[h]; ok {
			req.Header[h] = hh
		}
	}

	if frfToken != "" {
		req.Header.Set("X-Authentication-Token", frfToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		a.Log.WithField("error", err).Error("Can not execute proxy request")
		oAuthFail(H{"error": "Can not execute proxy request"}, w)
		return
	}

	defer resp.Body.Close()

	for _, h := range HeadersFromBackend {
		if hh, ok := resp.Header[h]; ok {
			w.Header()[h] = hh
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
