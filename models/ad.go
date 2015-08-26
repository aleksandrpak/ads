package models

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/bluele/gcache"
)

type AdsCollection interface {
	GetAdByID(adID bson.ObjectId) (*Ad, error)
	GetAdIDs(info *ClientInfo, startViewsCount int) (*[]ID, error)
	GetNewAd(info *ClientInfo, startViewsCount int) *Ad
}

type adsCollection struct {
	*mgo.Collection

	cache gcache.Cache
}

type Description struct {
	ID   bson.ObjectId `bson:"id" json:"-"`
	Text string        `bson:"text" json:"text"`
}

type Target struct {
	Geo     string  `bson:"geo" json:"geo"`
	Gender  string  `bson:"gender" json:"gender"`
	AgeLow  float64 `bson:"ageLow" json:"ageLow"`
	AgeHigh float64 `bson:"ageHigh" json:"ageHigh"`
}

type Ad struct {
	ID                  bson.ObjectId `bson:"_id" json:"-"`
	Name                string        `bson:"name" json:"name"`
	Campaign            bson.ObjectId `bson:"campaign" json:"campaign"`
	IsActive            bool          `bson:"isActive" json:"isActive"`
	IsCampaignActive    bool          `bson:"isCampaignActive" json:"isCampaignActive"`
	InstallURL          string        `bson:"installUrl" json:"installUrl"`
	IconURL             string        `bson:"iconUrl" json:"iconUrl"`
	FeedBannerURL       string        `bson:"feedBannerUrl" json:"feedBannerUrl"`
	FullScreenBannerURL string        `bson:"fullScreenBannerUrl" json:"fullScreenBannerUrl"`
	AdsBannerURL        string        `bson:"iAdsBannerUrl" json:"iAdsBannerUrl"`
	ViewsCount          int           `bson:"viewsCount" json:"viewsCount"`
	ShortDescriptions   []Description `bson:"shortDescriptions" json:"shortDescriptions"`
	LongDescriptions    []Description `bson:"longDescriptions" json:"longDescriptions"`
	Target              Target        `bson:"target" json:"target"`
}

type cacheKey struct {
	info            ClientInfo // TODO: accept only required fields for query
	startViewsCount int
}

func NewAdsCollection(c *mgo.Collection) AdsCollection {
	c.EnsureIndexKey(
		"isActive",
		"isCampaignActive",
		"viewsCount",
		"target.geo",
		"target.gender",
		"target.ageLow",
		"target.ageHigh",
	)

	return &adsCollection{c, gcache.
		New(100).
		Expiration(time.Second).
		LoaderFunc(func(key interface{}) (interface{}, error) {
		cKey := key.(cacheKey)

		var adIDs []ID
		query := buildQuery(&cKey.info, false, cKey.startViewsCount)
		err := c.Find(query).Sort("_id").Select(bson.M{"_id": 1}).All(&adIDs)

		return &adIDs, err

	}).Build()}
}

func (c *adsCollection) GetAdIDs(info *ClientInfo, startViewsCount int) (*[]ID, error) {
	adIDs, err := c.cache.Get(cacheKey{*info, startViewsCount})

	return adIDs.(*[]ID), err
}

func (c *adsCollection) GetAdByID(adID bson.ObjectId) (*Ad, error) {
	ad := &Ad{}
	_, err := c.FindId(adID).Apply(mgo.Change{Update: bson.M{"$inc": bson.M{"viewsCount": 1}}}, ad)
	return ad, err
}

func (c *adsCollection) GetNewAd(info *ClientInfo, startViewsCount int) *Ad {
	var ad Ad
	_, err := c.Find(buildQuery(info, true, startViewsCount)).Apply(mgo.Change{Update: bson.M{"$inc": bson.M{"viewsCount": 1}}}, &ad)
	if err != nil {
		return nil
	}

	return &ad
}

func buildQuery(info *ClientInfo, isNew bool, startViewsCount int) *bson.M {
	var trafficFilter interface{}

	if isNew {
		trafficFilter = bson.M{"viewsCount": bson.M{"$lte": startViewsCount}}
	} else {
		trafficFilter = bson.M{"viewsCount": bson.M{"$gt": startViewsCount}}
	}

	return &bson.M{
		"isActive":         true,
		"isCampaignActive": true,
		"$and": [5]interface{}{
			bson.M{"$or": getGeoQuery(info)},
			bson.M{"$or": getGenderQuery(info)},
			bson.M{"$or": getAgeLowQuery(info)},
			bson.M{"$or": getAgeHighQuery(info)},
			trafficFilter,
		},
	}
}

func getGeoQuery(info *ClientInfo) []interface{} {
	query := []interface{}{
		bson.M{"target.geo": bson.M{"$exists": false}},
	}

	if info.Geo.Country.ISOCode != "" {
		query = append(query, bson.M{"target.geo": info.Geo.Country.ISOCode})
	}

	return query
}

func getGenderQuery(info *ClientInfo) []interface{} {
	query := []interface{}{
		bson.M{"target.gender": bson.M{"$exists": false}},
	}

	if info.Gender != "" {
		query = append(query, bson.M{"target.gender": info.Gender})
	}

	return query
}

func getAgeLowQuery(info *ClientInfo) []interface{} {
	query := []interface{}{
		bson.M{"target.ageLow": bson.M{"$exists": false}},
	}

	if info.Age != 0 {
		query = append(query, bson.M{"target.ageLow": bson.M{"$lte": info.Age}})
	}

	return query
}

func getAgeHighQuery(info *ClientInfo) []interface{} {
	query := []interface{}{
		bson.M{"target.ageHigh": bson.M{"$exists": false}},
	}

	if info.Age != 0 {
		query = append(query, bson.M{"target.ageHigh": bson.M{"$gte": info.Age}})
	}

	return query
}
