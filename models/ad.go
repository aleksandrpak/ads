package models

import (
	"errors"
	"math/rand"
	"time"

	"github.com/aleksandrpak/ads/system/config"
	"github.com/aleksandrpak/ads/system/log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type AdsCollection interface {
	GetAd(client *Client, views ViewsCollection, conversions ConversionsCollection, dbConfig config.DbConfig) (*Ad, error)
}

type adsCollection struct {
	coll      *mgo.Collection
	isNewRand *rand.Rand
	rankRand  *rand.Rand
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

	return &adsCollection{
		coll:      c,
		isNewRand: rand.New(rand.NewSource(time.Now().Unix())),
		rankRand:  rand.New(rand.NewSource(time.Now().Unix()))}
}

func (c *adsCollection) GetAd(client *Client, views ViewsCollection, conversions ConversionsCollection, dbConfig config.DbConfig) (*Ad, error) {
	var ad *Ad
	startViewsCount := dbConfig.StartViewsCount()

	isNew := c.isNewRand.Float32() < dbConfig.NewTrafficPercentage()
	if isNew {
		ad := c.tryGetNewAd(client.Info, startViewsCount)
		if ad != nil {
			return ad, nil
		}
	}

	var adIDs []ID

	err := c.coll.Find(buildQuery(client.Info, false, startViewsCount)).Sort("_id").Select(bson.M{"_id": 1}).All(&adIDs)
	if err != nil {
		if !isNew {
			ad = c.tryGetNewAd(client.Info, startViewsCount)
		}

		if ad == nil {
			log.Error.Pf("failed to execute ad ids query: %v", err)
			return nil, errors.New("no ads found")
		} else {
			return ad, nil
		}
	}

	ids := make([]bson.ObjectId, len(adIDs))
	for i, id := range adIDs {
		ids[i] = id.ID
	}

	period := time.Now().UTC().Add(time.Duration(time.Hour) * time.Duration(dbConfig.StatisticHours()))

	viewsPerAd, err := views.GetStatistics(ids, period)
	if err != nil {
		log.Error.Pf("failed to get ads view statistics: %v", err)
		return nil, errors.New("failed to count view statistics")
	}

	conversionsPerAd, err := conversions.GetStatistics(ids, period)
	if err != nil {
		log.Error.Pf("failed to get ads conversion statistics: %v", err)
		return nil, errors.New("failed to count conversion statistics")
	}

	rankPerAd := make(map[bson.ObjectId]float32)
	totalRank := float32(0)
	for k, v := range *viewsPerAd {
		c, ok := (*conversionsPerAd)[k]
		var rank float32
		if ok {
			rank = c / v
		} else {
			rank = 1 / v
		}

		totalRank += rank
		rankPerAd[k] = rank
	}

	var adID bson.ObjectId
	currentWeight := float32(0)
	targetWeight := c.rankRand.Float32()
	for _, id := range ids {
		adID = id
		currentWeight += rankPerAd[adID] / totalRank
		if currentWeight >= targetWeight {
			break
		}
	}

	ad = &Ad{}
	_, err = c.coll.FindId(adID).Apply(mgo.Change{Update: bson.M{"$inc": bson.M{"viewsCount": 1}}}, ad)
	if err != nil {
		log.Error.Pf("failed to get ad: %v", err)
		return nil, errors.New("internal error while getting ad")
	}

	return ad, nil
}

func (c *adsCollection) tryGetNewAd(info *ClientInfo, startViewsCount int) *Ad {
	var ad Ad
	_, err := c.coll.Find(buildQuery(info, true, startViewsCount)).Apply(mgo.Change{Update: bson.M{"$inc": bson.M{"viewsCount": 1}}}, &ad)
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
