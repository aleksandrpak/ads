package models

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/bluele/gcache"
)

type AdsCollection interface {
	GetAdByID(adID *bson.ObjectId) (*Ad, error)
	GetAdIDs(info *ClientInfo, startViewsCount int) (*[]ID, error)
	GetNewAd(info *ClientInfo, startViewsCount int) *Ad
	ToggleAd(adID *bson.ObjectId, value bool)
}

type adsCollection struct {
	*mgo.Collection

	cache gcache.Cache
}

type Ad struct {
	ID                    bson.ObjectId `bson:"_id"`
	ConversionURL         string        `bson:"conversionUrl"`
	FeedBannerURL         string        `bson:"feedBannerUrl"`
	FeedDescription       string        `bson:"feedDescription"`
	FullscreenBannerURL   string        `bson:"fullscreenBannerUrl"`
	FullscreenDescription string        `bson:"fullscreenDescription"`
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
	if err != nil {
		return nil, err
	}

	return adIDs.(*[]ID), nil
}

func (c *adsCollection) GetAdByID(adID *bson.ObjectId) (*Ad, error) {
	ad := &Ad{}
	_, err := c.FindId(*adID).Apply(mgo.Change{Update: bson.M{"$inc": bson.M{"viewsCount": 1}}}, ad)
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

func (c *adsCollection) ToggleAd(adID *bson.ObjectId, value bool) {
	c.UpdateId(adID, bson.M{"$set": bson.M{"isActive": value}})
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
