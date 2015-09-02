package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"

	"git.startupteam.ru/aleksandrpak/ads/models"
	"git.startupteam.ru/aleksandrpak/ads/system/log"
	"github.com/julienschmidt/httprouter"
)

type AdsController interface {
	View(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Click(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Conversion(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}

func (c *controller) View(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	database := c.app.Database()
	apps := database.Apps()

	devApp, err := apps.GetApp(r)
	if err != nil {
		c.writeError(w, err)
		return
	}

	client, err := models.GetClient(c.app.GeoIP(), r)
	if err != nil {
		c.writeError(w, err)
		return
	}

	ad, err := c.strategy.NextAd(client)
	if err != nil {
		c.writeError(w, err)
		return
	}

	jsonAd, e := json.Marshal(ad)
	if err != nil {
		c.writeError(w, log.NewInternalError(e))
		return
	}

	go database.Views().SaveStatistic(ad.ID, devApp.ID, client)

	w.Write(jsonAd)
}

func (c *controller) Click(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	viewId := p.ByName("viewId")
	if viewId == "" {
		c.writeError(w, log.NewError(http.StatusBadRequest, "view id is not specified"))
		return
	}

	view, err := c.app.Database().Views().GetById(&viewId)
	if err != nil {
		c.writeError(w, err)
		return
	}

	c.app.Database().Clicks().SaveNextStatistic(view)

	d := func(req *http.Request) {
		req = r
		req.URL.Scheme = "http"
		req.URL.Host = "ya.ru"
		req.URL.Path = ""
		req.URL.RawQuery = ""
	}

	proxy := &httputil.ReverseProxy{Director: d}
	proxy.ServeHTTP(w, r)
}

func (c *controller) Conversion(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
}
