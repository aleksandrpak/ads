package main

import (
	"flag"
	"fmt"
	"net/http"
	"runtime"

	// TODO: Remove in production
	_ "expvar"
	_ "net/http/pprof"

	"github.com/golang/glog"

	"github.com/aleksandrpak/ads/controllers/api"
	"github.com/aleksandrpak/ads/system/application"
	"github.com/julienschmidt/httprouter"
)

func main() {
	configFilename := flag.String("config", "config.json", "Path to configuration file")

	runtime.GOMAXPROCS(runtime.NumCPU())

	app := application.NewApplication(configFilename)
	router := route(app)

	// TODO: Remove in production
	go func() {
		glog.Fatal(http.ListenAndServe("localhost:6060", nil))
	}()

	glog.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", app.AppConfig().Port()), router))
}

func route(app application.Application) *httprouter.Router {
	apiController := api.NewController(app)

	router := httprouter.New()

	router.GET("/api/ads", apiController.NextAd)

	return router
}
