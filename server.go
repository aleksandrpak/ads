package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	// TODO: Remove in production
	_ "expvar"
	_ "net/http/pprof"

	"github.com/aleksandrpak/ads/controllers/api"
	"github.com/aleksandrpak/ads/strategy/weightStrategy"
	"github.com/aleksandrpak/ads/system/application"
	"github.com/aleksandrpak/ads/system/log"
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

	log.Fatal.Pf("Failed to start: ", http.ListenAndServe(fmt.Sprintf(":%d", app.AppConfig().Port()), router))
}

func newApiController(app application.Application) api.Controller {
	strategy := weightStrategy.New(app.Database(), app.AppConfig().DbConfig())

	return api.NewController(app, strategy)
}

func route(apiController api.Controller) *httprouter.Router {
	router := httprouter.New()

	router.GET("/api/ads", apiController.NextAd)

	return router
}
