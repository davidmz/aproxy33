package app

import (
	"database/sql"

	"github.com/Sirupsen/logrus"
	"github.com/davidmz/aproxy33/codereg"
	"github.com/gorilla/mux"
)

type App struct {
	*Conf
	DB              *sql.DB
	Router          *mux.Router
	Log             *logrus.Logger
	LocalAuthTokens *codereg.Reg
	OAuthCodes      *codereg.Reg
}
