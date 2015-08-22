package models

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type ViewsCollection interface {
	SaveView(adID bson.ObjectId, appID bson.ObjectId, client *Client)
	GetStatistics(adIDs []bson.ObjectId, period time.Time) (*map[bson.ObjectId]float32, error)
}

type viewsCollection struct {
	*mgo.Collection
}

type View struct {
	AdID   bson.ObjectId `bson:"adId"`
	AppID  bson.ObjectId `bson:"appId"`
	Client Client        `bson:"client"`
	At     time.Time     `bson:"at"`
}

func NewViewsCollection(c *mgo.Collection) ViewsCollection {
	c.EnsureIndexKey(
		"adId",
		"at",
	)

	return &viewsCollection{c}
}

func (c *viewsCollection) SaveView(adID bson.ObjectId, appID bson.ObjectId, client *Client) {
	c.Insert(&View{
		AdID:   adID,
		AppID:  appID,
		Client: *client,
		At:     time.Now().UTC(),
	})
}

func (c *viewsCollection) GetStatistics(adIDs []bson.ObjectId, period time.Time) (*map[bson.ObjectId]float32, error) {
	views := c.Find(bson.M{"adId": bson.M{"$in": adIDs}, "at": bson.M{"$gte": period}}).Select(bson.M{"adId": 1}).Iter()
	viewsPerAd := make(map[bson.ObjectId]float32)

	var view View
	for views.Next(&view) {
		v, ok := viewsPerAd[view.AdID]
		if ok {
			viewsPerAd[view.AdID] = v + 1
		} else {
			viewsPerAd[view.AdID] = 1
		}
	}

	if err := views.Close(); err != nil {
		return nil, err
	}

	return &viewsPerAd, nil
}
