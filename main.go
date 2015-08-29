package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/davidmz/aproxy33/app"
	"github.com/gorilla/handlers"
	"github.com/justinas/alice"
	"github.com/rs/cors"
)

func main() {
	var confFile string
	flag.StringVar(&confFile, "c", "", "config file")
	flag.Parse()

	if confFile == "" {
		flag.Usage()
		os.Exit(0)
	}

	app := new(app.App)
	if err := app.LoadConfig(confFile); err != nil {
		log.Fatalf("Can not read config file: %v", err)
	}

	app.InitRouter()

	crs := cors.New(cors.Options{
		AllowedOrigins:   app.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"authorization"},
		AllowCredentials: true,
	})

	h := alice.New(
		LoggingHandler(os.Stdout),
		// app.CatchPanics,
		crs.Handler,
	).Then(app.Router)

	s := &http.Server{
		Addr:           app.Listen,
		Handler:        h,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	app.Log.Infof("Starting server at %s", app.Listen)
	app.Log.Fatal(s.ListenAndServe())
}

func LoggingHandler(out io.Writer) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return handlers.LoggingHandler(out, h)
	}
}
