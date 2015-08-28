package app

import (
	"database/sql"
	"encoding/json"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/davidmz/aproxy33/codereg"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Conf struct {
	Listen            string
	Secret            []byte
	DBUrl             string
	LogLevel          string
	CORSOrigins       []string
	BackendAPIRoot    string
	BackendAPIMethods map[string]string
}

func (a *App) LoadConfig(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	a.Conf = new(Conf)

	if err := json.NewDecoder(f).Decode(&a.Conf); err != nil {
		return err
	}

	a.DB, err = sql.Open("postgres", a.DBUrl)
	if err != nil {
		return err
	}

	if err := a.DB.Ping(); err != nil {
		return err
	}

	a.Router = mux.NewRouter()

	a.Log = logrus.New()
	a.Log.Out = os.Stderr
	a.Log.Formatter = &logrus.TextFormatter{ForceColors: true}
	a.Log.Level = logrus.ErrorLevel
	if ll, err := logrus.ParseLevel(a.LogLevel); err == nil {
		a.Log.Level = ll
	}

	a.LocalAuthTokens = codereg.New(20, 5*time.Minute, time.Minute)
	a.OAuthCodes = codereg.New(20, 10*time.Minute, time.Minute)

	return nil
}
