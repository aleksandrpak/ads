package models

import (
	"math/rand"

	"github.com/golang/glog"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type AdsCollection interface {
	GetAd(client *Client) *Ad
	UpdateAd(ad *Ad)
}

type adsCollection struct {
	*mgo.Collection
}

type Description struct {
	ID   bson.ObjectId `bson:"id,omitempty" json:"-"`
	Text string        `bson:"text" json:"text"`
}

type Target struct {
	Geo     string  `bson:"geo" json:"geo"`
	Gender  string  `bson:"gender" json:"gender"`
	AgeLow  float64 `bson:"ageLow" json:"ageLow"`
	AgeHigh float64 `bson:"ageHigh" json:"ageHigh"`
}

type Ad struct {
	ID                     bson.ObjectId `bson:"_id,omitempty" json:"-"`
	Name                   string        `bson:"name" json:"name"`
	Campaign               bson.ObjectId `bson:"campaign" json:"campaign"`
	IsActive               bool          `bson:"isActive" json:"isActive"`
	IsCampaignActive       bool          `bson:"isCampaignActive" json:"isCampaignActive"`
	InstallURL             string        `bson:"installUrl" json:"installUrl"`
	IconURL                string        `bson:"iconUrl" json:"iconUrl"`
	FeedBannerURL          string        `bson:"feedBannerUrl" json:"feedBannerUrl"`
	FullScreenBannerURL    string        `bson:"fullScreenBannerUrl" json:"fullScreenBannerUrl"`
	AdsBannerURL           string        `bson:"iAdsBannerUrl" json:"iAdsBannerUrl"`
	InstallationPrice      float64       `bson:"installationPrice" json:"installationPrice"`
	BudgetLimit            float64       `bson:"budgetLimit" json:"budgetLimit"`
	ConversionsLimit       float64       `bson:"conversionsLimit" json:"conversionsLimit"`
	StartViewCounts        float64       `bson:"startViewCounts" json:"startViewCounts"`
	ApproxViewsCount       float64       `bson:"approxViewsCount" json:"approxViewsCount"`
	ApproxClicksCount      float64       `bson:"approxClicksCount" json:"approxClicksCount"`
	ApproxConversionsCount float64       `bson:"approxConversionsCount" json:"approxConversionsCount"`
	ApproxRank             float64       `bson:"approxRank" json:"approxRank"`
	ShortDescriptions      []Description `bson:"shortDescriptions" json:"shortDescriptions"`
	LongDescriptions       []Description `bson:"longDescriptions" json:"longDescriptions"`
	Target                 Target        `bson:"target" json:"target"`
}

func NewAdsCollection(c *mgo.Collection) AdsCollection {
	return &adsCollection{c}
}

func (c *adsCollection) GetAd(client *Client) *Ad {
	ads := c.Find(bson.M{
		"isActive":         true,
		"isCampaignActive": true,
		"$or": [4]interface{}{
			bson.M{"$or": [2]interface{}{bson.M{"target.geo": "Global"}, bson.M{"target.geo": "Russia"}}},
			bson.M{"$or": [2]interface{}{bson.M{"target.gender": nil}, bson.M{"target.gender": "female"}}},
			bson.M{"$or": [2]interface{}{bson.M{"target.ageLow": -1}, bson.M{"target.ageLow": bson.M{"$lte": client.Age}}}},
			bson.M{"$or": [2]interface{}{bson.M{"target.ageHigh": -1}, bson.M{"target.ageHigh": bson.M{"$gte": client.Age}}}},
		}}).Sort("-approxRank").Select(bson.M{"_id": 1}).Limit(rand.Intn(101) + 1).Iter()

	var adID struct {
		ID bson.ObjectId `bson:"_id,omitempty"`
	}

	for ads.Next(&adID) {
	}

	if err := ads.Close(); err != nil {
		glog.Errorf("failed to execute ads query: %v", err)
		return nil
	}

	var ad Ad
	err := c.FindId(adID.ID).One(&ad)
	if err != nil {
		glog.Errorf("failed to get ad: %v", err)
		return nil
	}

	return &ad
}

func (c *adsCollection) UpdateAd(ad *Ad) {
	viewsCount := ad.ApproxViewsCount + 1
	rank := 1.0

	if viewsCount > ad.StartViewCounts {
		rank = ad.ApproxConversionsCount / viewsCount
	}

	c.UpdateId(ad.ID, bson.M{"$inc": bson.M{"approxViewsCount": 1}, "approxRank": rank})
}
