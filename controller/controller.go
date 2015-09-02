package controller

import (
	"fmt"
	"net/http"

	"git.startupteam.ru/aleksandrpak/ads/strategy"
	"git.startupteam.ru/aleksandrpak/ads/system/application"
	"git.startupteam.ru/aleksandrpak/ads/system/log"
)

type Controller interface {
	AdsController
}

type controller struct {
	app      application.Application
	strategy strategy.Strategy
}

func NewController(app application.Application, strategy strategy.Strategy) Controller {
	return &controller{app, strategy}
}

func (c *controller) writeError(w http.ResponseWriter, err log.ServerError) {
	log.Error.Er(err)

	desc := err.Desc()
	if desc == nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"error\":\"internal server error\"}"))
	} else {
		w.WriteHeader(err.Status())
		w.Write([]byte(fmt.Sprintf("{\"error\":\"%v\"}", *desc)))
	}
}
