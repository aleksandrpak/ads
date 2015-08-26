package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	// TODO: Remove in production
	_ "expvar"
	_ "net/http/pprof"

	"git.startupteam.ru/aleksandrpak/ads/controllers/api"
	"git.startupteam.ru/aleksandrpak/ads/strategy/rankStrategy"
	"git.startupteam.ru/aleksandrpak/ads/system/application"
	"git.startupteam.ru/aleksandrpak/ads/system/log"
	"github.com/julienschmidt/httprouter"
)

func main() {
	configFilename := flag.String("config", "config.json", "Path to configuration file")

	// TODO: Add different destinations
	log.Init(os.Stderr, os.Stderr, os.Stdout, os.Stdout, os.Stdout)

	app := application.NewApplication(configFilename)
	apiController := newApiController(app)
	router := route(apiController)

	// TODO: Remove in production
	go func() {
		log.Fatal.Pf("Failed to start profiling tools: ", http.ListenAndServe("localhost:6060", nil))
	}()

	// TODO: Make graceful shutdown
	log.Fatal.Pf("Failed to start: ", http.ListenAndServe(fmt.Sprintf(":%d", app.AppConfig().Port()), router))
}

func newApiController(app application.Application) api.Controller {
	strategy := rankStrategy.New(app.Database(), app.AppConfig().DbConfig())

	return api.NewController(app, strategy)
}

func route(apiController api.Controller) *httprouter.Router {
	router := httprouter.New()

	router.GET("/api/ads", apiController.NextAd)

	return router
}
