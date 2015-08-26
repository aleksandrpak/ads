package models

import (
	"errors"
	"net/http"

	"git.startupteam.ru/aleksandrpak/ads/system/log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type AppsCollection interface {
	GetApp(r *http.Request) (*App, error)
}

type appsCollection struct {
	*mgo.Collection
}

type App struct {
	ID       bson.ObjectId `bson:"_id" json:"-"`
	AppToken string        `bson:"appToken"`
}

func NewAppsCollection(c *mgo.Collection) AppsCollection {
	c.EnsureIndexKey(
		"appToken",
	)

	return &appsCollection{c}
}

func (c *appsCollection) GetApp(r *http.Request) (*App, error) {
	token := r.URL.Query().Get("appToken")
	if token == "" {
		return nil, errors.New("App token is not specified")
	}

	var app App
	err := c.Find(&bson.M{"appToken": token}).One(&app)
	if err != nil {
		log.Error.Pf("failed to get app with token \"%v\": %v", token, err)
		return nil, errors.New("App token is not registered")
	}

	return &app, nil
}
