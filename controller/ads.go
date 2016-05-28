package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"gopkg.in/mgo.v2/bson"

	"github.com/aleksandrpak/ads/models"
	"github.com/aleksandrpak/ads/models/statistic"
	"github.com/aleksandrpak/ads/system/database"
	"github.com/aleksandrpak/ads/system/log"
	"github.com/julienschmidt/httprouter"
)

const (
	feed       string = "feed"
	fullscreen string = "fullscreen"
)

type adInfo struct {
	ActionURL   string `json:"actionUrl"`
	BannerURL   string `json:"bannerUrl"`
	Description string `json:"description"`
}

type AdsController interface {
	View(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Click(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Conversion(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}

func (c *controller) View(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	database := c.app.Database()
	apps := database.Apps()

	t := getAdType(r)
	if t == "" {
		c.writeError(w, log.NewError(http.StatusBadRequest, "type of request ad is not specified"))
		return
	}

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

	viewId := database.Views().SaveStatistic(ad.ID, devApp.ID, client)
	adInfo := getInfo(t, viewId, ad, r)

	jsonAd, e := json.Marshal(adInfo)
	if e != nil {
		c.writeError(w, log.NewInternalError(e))
		return
	}

	w.Write(jsonAd)
}

func (c *controller) Click(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	view, err := getStatistic(p.ByName("viewId"), c.app.Database().Views())
	if err != nil {
		c.writeError(w, err)
		return
	}

	clickID := c.app.Database().Clicks().SaveNextStatistic(view)
	ad, e := c.app.Database().Ads().GetAdByID(&view.AdID)
	if e != nil {
		c.writeError(w, log.NewInternalError(e))
		return
	}

	conversion, err := c.app.Database().Conversions().GetLast(&view.AdID, 1)
	if err == nil {
		click, err := c.app.Database().Clicks().GetLast(&view.AdID, 100)
		if err == nil && click != nil && (conversion == nil || click.At > conversion.At) {
			toggleAd(&view.AdID, c.app.Database(), false)
		}
	}

	url, e := url.ParseRequestURI(ad.ConversionURL + clickID.Hex())
	if e != nil {
		c.writeError(w, log.NewInternalError(e))
		return
	}

	d := func(req *http.Request) {
		req = r
		req.URL.Scheme = url.Scheme
		req.URL.Host = url.Host
		req.URL.Path = url.Path
		req.URL.RawQuery = url.RawQuery
	}

	proxy := &httputil.ReverseProxy{Director: d}
	proxy.ServeHTTP(w, r)
}

func (c *controller) Conversion(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	click, err := getStatistic(p.ByName("clickId"), c.app.Database().Clicks())
	if err != nil {
		c.writeError(w, err)
		return
	}

	toggleAd(&click.AdID, c.app.Database(), true)
	c.app.Database().Clicks().SaveNextStatistic(click)
}

func toggleAd(adID *bson.ObjectId, d database.Database, value bool) {
	conversions := d.Conversions().GetStatisticCount(adID)
	if conversions == 0 {
		d.Ads().ToggleAd(adID, value)
	}
}

func getStatistic(id string, c statistic.StatisticsCollection) (*statistic.Statistic, log.ServerError) {
	if id == "" {
		return nil, log.NewError(http.StatusBadRequest, "id is not specified")
	}

	return c.GetById(&id)
}

func getInfo(t string, viewId *bson.ObjectId, ad *models.Ad, r *http.Request) *adInfo {
	scheme := r.URL.Scheme
	if scheme == "" {
		scheme = "http"
	}

	actionURL := fmt.Sprintf("%v://%v/ads/click/%v", scheme, r.Host, viewId.Hex())

	switch t {
	case feed:
		return &adInfo{
			ActionURL:   actionURL,
			BannerURL:   ad.FeedBannerURL,
			Description: ad.FeedDescription,
		}

	case fullscreen:
		return &adInfo{
			ActionURL:   actionURL,
			BannerURL:   ad.FullscreenBannerURL,
			Description: ad.FullscreenDescription,
		}
	}

	return nil
}

func getAdType(r *http.Request) string {
	switch r.URL.Query().Get("type") {
	case feed:
		return feed
	case fullscreen:
		return fullscreen
	}

	return ""
}
