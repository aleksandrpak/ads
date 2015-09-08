package database

import (
	"git.startupteam.ru/aleksandrpak/ads/models"
	"git.startupteam.ru/aleksandrpak/ads/models/statistic"
	"git.startupteam.ru/aleksandrpak/ads/system/config"
	"git.startupteam.ru/aleksandrpak/ads/system/log"
	mgo "gopkg.in/mgo.v2"
)

type Database interface {
	Ads() models.AdsCollection
	Apps() models.AppsCollection
	Views() statistic.StatisticsCollection
	Clicks() statistic.StatisticsCollection
	Conversions() statistic.StatisticsCollection
	Close()
}

type database struct {
	session     *mgo.Session
	ads         models.AdsCollection
	apps        models.AppsCollection
	views       statistic.StatisticsCollection
	clicks      statistic.StatisticsCollection
	conversions statistic.StatisticsCollection
}

func Connect(dbConfig config.DbConfig) Database {
	session, err := mgo.Dial(dbConfig.Hosts())

	if err != nil {
		log.Fatal.Pf("Can't connect to the database: %v", err)
		panic(err)
	}

	session.SetMode(mgo.Eventual, true)

	db := session.DB(dbConfig.Database())

	return &database{
		session:     session,
		ads:         models.NewAdsCollection(db.C("ads")),
		apps:        models.NewAppsCollection(db.C("apps")),
		views:       statistic.NewStatisticsCollection(db.C("views"), dbConfig.StatisticHours()),
		clicks:      statistic.NewStatisticsCollection(db.C("clicks"), dbConfig.StatisticHours()),
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

func (d *database) Clicks() statistic.StatisticsCollection {
	return d.clicks
}

func (d *database) Conversions() statistic.StatisticsCollection {
	return d.conversions
}

func (d *database) Close() {
	d.session.Close()
}
