package database

import (
	"github.com/aleksandrpak/ads/models"
	"github.com/aleksandrpak/ads/models/statistic"
	"github.com/aleksandrpak/ads/system/config"
	"github.com/aleksandrpak/ads/system/log"
	mgo "gopkg.in/mgo.v2"
)

type Database interface {
	Ads() models.AdsCollection
	Apps() models.AppsCollection
	Views() statistic.StatisticsCollection
	Conversions() statistic.StatisticsCollection
}

type database struct {
	// TODO: save session and close on app exit
	ads         models.AdsCollection
	apps        models.AppsCollection
	views       statistic.StatisticsCollection
	conversions statistic.StatisticsCollection
}

func Connect(dbConfig config.DbConfig) Database {
	dbSession, err := mgo.Dial(dbConfig.Hosts())

	if err != nil {
		log.Fatal.Pf("Can't connect to the database: %v", err)
		panic(err)
	}

	dbSession.SetMode(mgo.Eventual, true)

	db := dbSession.DB(dbConfig.Database())

	return &database{
		ads:         models.NewAdsCollection(db.C("ads")),
		apps:        models.NewAppsCollection(db.C("apps")),
		views:       statistic.NewStatisticsCollection(db.C("views"), dbConfig.StatisticHours()),
		conversions: statistic.NewStatisticsCollection(db.C("conversions"), dbConfig.StatisticHours()),
	}
}

func (d *database) Ads() models.AdsCollection {
	return d.ads
}

func (d *database) Apps() models.AppsCollection {
	return d.apps
}

func (d *database) Views() statistic.StatisticsCollection {
	return d.views
}

func (d *database) Conversions() statistic.StatisticsCollection {
	return d.conversions
}
