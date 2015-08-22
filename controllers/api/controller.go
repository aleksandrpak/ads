package api

import (
	"fmt"
	"net/http"

	"github.com/aleksandrpak/ads/system/application"
)

type Controller interface {
	AdsController
}

type controller struct {
	app application.Application
}

func NewController(app application.Application) Controller {
	return &controller{app}
}

func (c *controller) writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf("{\"error\":\"%v\"}", err)))
}
