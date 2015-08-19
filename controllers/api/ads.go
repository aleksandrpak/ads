package api

import (
	"encoding/json"
	"net/http"

	"github.com/aleksandrpak/ads/models"
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
)

type AdsController interface {
	NextAd(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}

func (c *controller) NextAd(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	client := models.GetClient(c.app.GeoIP(), r)
	ads := c.app.Database().Ads()

	ad := ads.GetAd(client)
	if ad == nil {
		return
	}

	go ads.UpdateAd(ad)

	jsonAd, err := json.Marshal(ad)
	if err != nil {
		glog.Errorf("failed to serialize to json: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonAd)
}
