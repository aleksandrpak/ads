package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	// TODO: Remove in production
	//_ "expvar"
	//_ "net/http/pprof"

	"git.startupteam.ru/aleksandrpak/ads/controller"
	"git.startupteam.ru/aleksandrpak/ads/strategy/rankStrategy"
	"git.startupteam.ru/aleksandrpak/ads/system/application"
	"git.startupteam.ru/aleksandrpak/ads/system/log"
	"github.com/julienschmidt/httprouter"
)

func main() {
	configFilename := flag.String("config", "config.json", "Path to configuration file")

	log.Init(os.Stderr, os.Stdout, os.Stdout, os.Stdout)

	app := application.NewApplication(configFilename)
	controller := newController(app)
	router := route(controller)

	// TODO: Remove in production
	// go func() {
	// 	log.Fatal.Pf("Failed to start profiling tools: ", http.ListenAndServe("localhost:6060", nil))
	// }()

	// TODO: Make graceful shutdown
	log.Fatal.Pf("Failed to start: ", http.ListenAndServe(fmt.Sprintf(":%d", app.AppConfig().Port()), router))
}

func newController(app application.Application) controller.Controller {
	strategy := rankStrategy.New(app.Database(), app.AppConfig().DbConfig())

	return controller.NewController(app, strategy)
}

func route(c controller.Controller) *httprouter.Router {
	router := httprouter.New()

	router.GET("/ads/view", c.View)
	router.GET("/ads/click/:viewId", c.Click)
	router.GET("/ads/conversion/:clickId", c.Conversion)

	return router
}
