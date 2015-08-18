package api

import "github.com/aleksandrpak/ads/system/application"

type Controller interface {
	AdsController
}

type controller struct {
	app application.Application
}

func NewController(app application.Application) Controller {
	return &controller{app}
}
