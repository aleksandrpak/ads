package models

import (
	"fmt"
	"net/http"

	"git.startupteam.ru/aleksandrpak/ads/system/log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type AppsCollection interface {
	GetApp(r *http.Request) (*App, log.ServerError)
}

type appsCollection struct {
	*mgo.Collection
}

type App struct {
	ID       bson.ObjectId `bson:"_id"`
	AppToken string        `bson:"appToken"`
}

func NewAppsCollection(c *mgo.Collection) AppsCollection {
	c.EnsureIndexKey(
		"appToken",
	)

	return &appsCollection{c}
}

func (c *appsCollection) GetApp(r *http.Request) (*App, log.ServerError) {
	token := r.URL.Query().Get("appToken")
	if token == "" {
		return nil, log.NewError(http.StatusBadRequest, "App token is not specified")
	}

	var app App
	err := c.Find(&bson.M{"appToken": token}).One(&app)
	if err != nil {
		return nil, log.New(http.StatusForbidden, fmt.Sprintf("app token %v is not found", token), err)
	}

	return &app, nil
}
