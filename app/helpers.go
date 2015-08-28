package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
)

func mustBeOK(err error) {
	if err != nil {
		panic(err)
	}
}

func mustBeOKVal(val interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return val
}

func mustBeOKOr(err error, errs ...error) error {
	if err == nil {
		return nil
	}
	for _, e := range errs {
		if e == err {
			return err
		}
	}
	panic(err)
}

type H map[string]interface{}

type ApiHandler func(r *http.Request) (httpCode int, result interface{})

func (a *App) ApiCall(ah ApiHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var (
			httpCode int
			result   interface{}
		)

		func() {
			defer func() {
				// handle panics
				if rec := recover(); rec != nil {
					a.Log.WithField("panic", rec).Error("panic happens")
					if a.Log.Level >= logrus.DebugLevel {
						buf := make([]byte, 1024)
						n := runtime.Stack(buf, false)
						a.Log.WithField("stack", string(buf[:n])).Debug("panic happens")
					}
					httpCode, result = http.StatusInternalServerError, "Internal error"
				}
			}()

			httpCode, result = ah(r)
		}()

		jResult := H{"status": "ok", "data": result}
		if httpCode/100 == http.StatusBadRequest/100 {
			jResult["status"] = "error"
		}
		if httpCode/100 == http.StatusInternalServerError/100 {
			jResult["status"] = "fail"
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpCode)
		json.NewEncoder(w).Encode(jResult)
	}
}

func (a *App) LocalAuthRequired(h ApiHandler) ApiHandler {
	return func(r *http.Request) (int, interface{}) {
		auh := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(auh) != 2 || auh[0] != "X-AProxy" {
			return http.StatusUnauthorized, "Not authorized"
		}

		vUserID, found := a.LocalAuthTokens.Get(auh[1])
		if !found {
			return http.StatusForbidden, "Not authorized"
		}

		context.Set(r, "UserID", vUserID)
		return h(r)
	}
}

// https://gist.github.com/swdunlop/9629168
func identifyPanic() string {
	var name, file string
	var line int
	var pc [16]uintptr

	n := runtime.Callers(3, pc[:])
	for _, pc := range pc[:n] {
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		file, line = fn.FileLine(pc)
		name = fn.Name()
		if !strings.HasPrefix(name, "runtime.") {
			break
		}
	}

	switch {
	case name != "":
		return fmt.Sprintf("%v:%v", name, line)
	case file != "":
		return fmt.Sprintf("%v:%v", file, line)
	}

	return fmt.Sprintf("pc:%x", pc)
}
