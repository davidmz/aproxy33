package app

func (app *App) InitRouter() {
	app.Router.HandleFunc("/aprox/auth", app.ApiCall(app.AuthInit)).Methods("POST")
	app.Router.HandleFunc("/aprox/auth-refresh", app.ApiCall(app.LocalAuthRequired(app.AuthRefresh))).Methods("POST")

	app.Router.HandleFunc("/aprox/apps", app.ApiCall(app.LocalAuthRequired(app.AppsList))).Methods("GET")
	app.Router.HandleFunc("/aprox/apps", app.ApiCall(app.LocalAuthRequired(app.AppsNew))).Methods("POST")
	app.Router.HandleFunc("/aprox/apps/{key}", app.ApiCall(app.LocalAuthRequired(app.AppsUpdate))).Methods("PUT")
	app.Router.HandleFunc("/aprox/apps/{key}", app.ApiCall(app.LocalAuthRequired(app.AppsDelete))).Methods("DELETE")
	app.Router.HandleFunc("/aprox/apps/{key}/info", app.ApiCall(app.AppsInfo)).Methods("GET")

	app.Router.HandleFunc("/aprox/tokens", app.ApiCall(app.LocalAuthRequired(app.TokensList))).Methods("GET")
	app.Router.HandleFunc("/aprox/tokens/{id}", app.ApiCall(app.LocalAuthRequired(app.TokensDelete))).Methods("DELETE")

	app.Router.HandleFunc("/aprox/oauth-code", app.ApiCall(app.LocalAuthRequired(app.OAuthNewCode))).Methods("POST")
	app.Router.HandleFunc("/oauth/token", app.OAuthToken).Methods("POST")

	app.Router.PathPrefix("/frf/").HandlerFunc(app.ProxyBackend)
}
