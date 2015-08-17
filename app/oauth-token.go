package app

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Выдача токена по авторизационному коду
// Ожидает POST-поля: "code", "client_id", "client_secret", "grant_type" = "authorization_code", "redirect_uri"
func (a *App) OAuthToken(w http.ResponseWriter, r *http.Request) {
	defer func() {
		// handle panics
		if rec := recover(); rec != nil {
			a.Log.WithField("panic", rec).Error("panic happens")
			oAuthFail(H{"error": "invalid_request", "error_description": "Internal error"}, w)
		}
	}()

	if err := r.ParseForm(); err != nil {
		oAuthError(H{"error": "invalid_request", "error_description": "Can not parse POST data"}, w)
		return
	}

	requiredFields := []string{"code", "client_id", "client_secret", "grant_type", "redirect_uri"}
	fields := map[string]string{}
	errStrings := []string{}
	for _, f := range requiredFields {
		if len(r.PostForm[f]) > 1 {
			errStrings = append(errStrings, "more than one '"+f+"' field")
		} else if len(r.PostForm[f]) == 0 || r.PostForm.Get(f) == "" {
			errStrings = append(errStrings, "empty '"+f+"' field")
		} else {
			fields[f] = r.PostForm.Get(f)
		}
	}
	if len(errStrings) > 0 {
		oAuthError(H{"error": "invalid_request", "error_description": "Invalid data: " + strings.Join(errStrings, "; ")}, w)
		return
	}
	if fields["grant_type"] != "authorization_code" {
		oAuthError(H{"error": "invalid_request", "error_description": "Invalid grant_type"}, w)
		return
	}

	var codeInfo *OAuthCodeInfo

	if v, found := a.OAuthCodes.Get(fields["code"]); !found {
		oAuthError(H{"error": "invalid_grant", "error_description": "Authorization code not found"}, w)
		return
	} else {
		codeInfo = v.(*OAuthCodeInfo)
	}

	if codeInfo.RedirectURI != fields["redirect_uri"] {
		oAuthError(H{"error": "invalid_request", "error_description": "'redirect_uri' not matches with code request"}, w)
		return
	}

	var appKey, appSecret string

	err := a.DB.QueryRow("select key, secret from "+a.DBTablePrefix+"apps where id = $1", codeInfo.AppID).Scan(&appKey, &appSecret)
	if mustBeOKOr(err, sql.ErrNoRows) != nil {
		oAuthError(H{"error": "invalid_request", "error_description": "App not found"}, w)
		return
	}

	if appKey != fields["client_id"] || appSecret != fields["client_secret"] {
		oAuthError(H{"error": "invalid_client", "error_description": "Client authentication failed"}, w)
		return
	}

	// всё в порядке

	b := make([]byte, 20)
	mustBeOKVal(rand.Read(b))
	accessToken := fmt.Sprintf("%x", b)

	// Сохраняем

	mustBeOKVal(a.DB.Exec(
		"insert into "+a.DBTablePrefix+"atokens (app_id, user_id, token, perms) values ($1, $2, $3, $4)",
		codeInfo.AppID, codeInfo.UserID, accessToken, codeInfo.Perms,
	))

	// для удобства сразу выдаём username
	username := ""
	mustBeOK(a.DB.QueryRow("select username from "+a.DBTablePrefix+"users where id = $1", codeInfo.UserID).Scan(&username))

	a.OAuthCodes.Del(fields["code"])

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(H{
		"access_token": accessToken,
		"token_type":   "bearer",
		"username":     username,
	})
}

func oAuthError(eData H, w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(eData)
}

func oAuthFail(eData H, w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(eData)
}
