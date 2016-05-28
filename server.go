package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "expvar"
	_ "net/http/pprof"

	"github.com/aleksandrpak/ads/controller"
	"github.com/aleksandrpak/ads/strategy/rankStrategy"
	"github.com/aleksandrpak/ads/system/application"
	"github.com/aleksandrpak/ads/system/log"
	"github.com/braintree/manners"
	"github.com/julienschmidt/httprouter"
)

func main() {
	configFilename := flag.String("config", "config.json", "Path to configuration file")

	log.Init(os.Stderr, os.Stdout, os.Stdout, os.Stdout)

	app := application.NewApplication(configFilename)
	controller := newController(app)
	router := route(controller)

	launchProf()
	waitShutdown(app)

	log.Fatal.Pf("failed to start: ", manners.ListenAndServe(fmt.Sprintf(":%d", app.AppConfig().Port()), router))
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

func launchProf() {
	go func() {
		log.Fatal.Pf("failed to start profiling tools: ", http.ListenAndServe("localhost:6060", nil))
	}()
}

func waitShutdown(app application.Application) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println()
		log.Info.Pf("gracefully shutting down")

		log.Info.Pf("stopping http listener")
		manners.Close()
		log.Info.Pf("http listener stopped")

		app.Cleanup()

		log.Info.Pf("shutdown finished")
		os.Exit(1)
	}()
}
