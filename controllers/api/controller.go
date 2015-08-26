package api

import (
	"fmt"
	"net/http"

	"git.startupteam.ru/aleksandrpak/ads/strategy"
	"git.startupteam.ru/aleksandrpak/ads/system/application"
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

func (c *controller) writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf("{\"error\":\"%v\"}", err)))
}
