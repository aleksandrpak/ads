package models

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type ConversionsCollection interface {
	GetStatistics(adIDs []bson.ObjectId, period time.Time) (*map[bson.ObjectId]float32, error)
}

type conversionsCollection struct {
	*mgo.Collection
}

type Conversion struct {
	AdID   bson.ObjectId `bson:"adId"`
	AppID  bson.ObjectId `bson:"appId"`
	Client Client        `bson:"client"`
	At     time.Time     `bson:"at"`
}

func NewConversionsCollection(c *mgo.Collection) ConversionsCollection {
	c.EnsureIndexKey(
		"adId",
		"at",
	)

	return &conversionsCollection{c}
}

func (c *conversionsCollection) GetStatistics(adIDs []bson.ObjectId, period time.Time) (*map[bson.ObjectId]float32, error) {
	conversions := c.Find(bson.M{"adId": bson.M{"$in": adIDs}, "at": bson.M{"$gte": period}}).Select(bson.M{"adId": 1}).Iter()
	conversionsPerAd := make(map[bson.ObjectId]float32)

	var conversion Conversion
	for conversions.Next(&conversion) {
		v, ok := conversionsPerAd[conversion.AdID]
		if ok {
			conversionsPerAd[conversion.AdID] = v + 1
		} else {
			conversionsPerAd[conversion.AdID] = 1
		}
	}

	if err := conversions.Close(); err != nil {
		return nil, err
	}

	return &conversionsPerAd, nil
}
