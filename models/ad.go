package models

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type AdsCollection interface {
	GetAdByID(adID bson.ObjectId) (*Ad, error)
	GetAdIDs(info *ClientInfo, startViewsCount int) (*[]bson.ObjectId, error)
	GetNewAd(info *ClientInfo, startViewsCount int) *Ad
}

type adsCollection struct {
	*mgo.Collection
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

	return &adsCollection{c}
}

func (c *adsCollection) GetAdIDs(info *ClientInfo, startViewsCount int) (*[]bson.ObjectId, error) {
	var adIDs []ID
	err := c.Find(buildQuery(info, false, startViewsCount)).Sort("_id").Select(bson.M{"_id": 1}).All(&adIDs)
	if err != nil {
		return nil, err
	}

	ids := make([]bson.ObjectId, len(adIDs))
	for i, id := range adIDs {
		ids[i] = id.ID
	}

	return &ids, err
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
		trafficFilter = bson.M{"viewsCount": bson.M{"$lt": startViewsCount}}
	} else {
		trafficFilter = bson.M{"viewsCount": bson.M{"$gte": startViewsCount}}
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
