package api

import (
	"encoding/json"
	"net/http"

	"github.com/aleksandrpak/ads/models"
	"github.com/aleksandrpak/ads/system/log"
	"github.com/julienschmidt/httprouter"
)

type AdsController interface {
	NextAd(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}

func (c *controller) NextAd(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	database := c.app.Database()
	apps := database.Apps()

	app, err := apps.GetApp(r)
	if err != nil {
		log.Error.Pf("failed to get app: %v", err)
		c.writeError(w, err)
		return
	}

	client, err := models.GetClient(c.app.GeoIP(), r)
	if err != nil {
		log.Error.Pf("failed to get client: %v", err)
		c.writeError(w, err)
		return
	}

	views := database.Views()
	ad, err := database.Ads().GetAd(client, views, database.Conversions(), c.app.AppConfig().DbConfig())
	if err != nil {
		log.Error.Pf("failed to get ad: %v", err)
		c.writeError(w, err)
		return
	}

	jsonAd, err := json.Marshal(ad)
	if err != nil {
		log.Error.Pf("failed to serialize ad to json: %v", err)
		c.writeError(w, err)
		return
	}

	go views.SaveView(ad.ID, app.ID, client)

	w.Write(jsonAd)
}
