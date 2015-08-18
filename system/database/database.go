package database

import (
	"github.com/aleksandrpak/ads/models"
	"github.com/aleksandrpak/ads/system/config"
	"github.com/golang/glog"
	mgo "gopkg.in/mgo.v2"
)

type Database interface {
	Ads() models.AdsCollection
}

type database struct {
	// TODO: save session and close on app exit
	ads models.AdsCollection
}

func Connect(dbConfig config.DbConfig) Database {
	dbSession, err := mgo.Dial(dbConfig.Hosts())

	if err != nil {
		glog.Fatalf("Can't connect to the database: %v", err)
		panic(err)
	}

	dbSession.SetMode(mgo.Eventual, true)

	db := dbSession.DB(dbConfig.Database())

	return &database{ads: models.NewAdsCollection(db.C("ads"))}
}

func (d *database) Ads() models.AdsCollection {
	return d.ads
}
