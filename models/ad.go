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

type targetInfo struct {
	Geo             string
	Gender          string
	Age             byte
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
		var adIDs []ID
		info := key.(targetInfo)
		err := c.Find(buildQuery(&info, false)).Sort("_id").Select(bson.M{"_id": 1}).All(&adIDs)
		return &adIDs, err
	}).Build()}
}

func (c *adsCollection) GetAdIDs(info *ClientInfo, startViewsCount int) (*[]ID, error) {
	adIDs, err := c.cache.Get(targetInfo{info.Geo.Country.ISOCode, info.Gender, info.Age, startViewsCount})
	return adIDs.(*[]ID), err
}

func (c *adsCollection) GetAdByID(adID bson.ObjectId) (*Ad, error) {
	ad := &Ad{}
	_, err := c.FindId(adID).Apply(mgo.Change{Update: bson.M{"$inc": bson.M{"viewsCount": 1}}}, ad)
	return ad, err
}

func (c *adsCollection) GetNewAd(info *ClientInfo, startViewsCount int) *Ad {
	var ad Ad
	_, err := c.
		Find(buildQuery(&targetInfo{info.Geo.Country.ISOCode, info.Gender, info.Age, startViewsCount}, true)).
		Apply(mgo.Change{Update: bson.M{"$inc": bson.M{"viewsCount": 1}}}, &ad)

	if err != nil {
		return nil
	}

	return &ad
}

func buildQuery(info *targetInfo, isNew bool) *bson.M {
	var trafficFilter interface{}

	if isNew {
		trafficFilter = bson.M{"viewsCount": bson.M{"$lte": info.startViewsCount}}
	} else {
		trafficFilter = bson.M{"viewsCount": bson.M{"$gt": info.startViewsCount}}
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

func getGeoQuery(info *targetInfo) []interface{} {
	query := []interface{}{
		bson.M{"target.geo": bson.M{"$exists": false}},
	}

	if info.Geo != "" {
		query = append(query, bson.M{"target.geo": info.Geo})
	}

	return query
}

func getGenderQuery(info *targetInfo) []interface{} {
	query := []interface{}{
		bson.M{"target.gender": bson.M{"$exists": false}},
	}

	if info.Gender != "" {
		query = append(query, bson.M{"target.gender": info.Gender})
	}

	return query
}

func getAgeLowQuery(info *targetInfo) []interface{} {
	query := []interface{}{
		bson.M{"target.ageLow": bson.M{"$exists": false}},
	}

	if info.Age != 0 {
		query = append(query, bson.M{"target.ageLow": bson.M{"$lte": info.Age}})
	}

	return query
}

func getAgeHighQuery(info *targetInfo) []interface{} {
	query := []interface{}{
		bson.M{"target.ageHigh": bson.M{"$exists": false}},
	}

	if info.Age != 0 {
		query = append(query, bson.M{"target.ageHigh": bson.M{"$gte": info.Age}})
	}

	return query
}
