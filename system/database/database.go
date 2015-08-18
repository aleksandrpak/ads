package database

import (
	"github.com/aleksandrpak/ads/system/config"
	"github.com/golang/glog"
	mgo "gopkg.in/mgo.v2"
)

type Database interface {
	Ads() *mgo.Collection
}

type database struct {
	// TODO: save session and close on app exit
	*mgo.Database
}

func Connect(dbConfig config.DbConfig) Database {
	dbSession, err := mgo.Dial(dbConfig.Hosts())

	if err != nil {
		glog.Fatalf("Can't connect to the database: %v", err)
		panic(err)
	}

	dbSession.SetMode(mgo.Eventual, true)

	return &database{dbSession.DB(dbConfig.Database())}
}

func (d *database) Ads() *mgo.Collection {
	return d.C("ads")
}
