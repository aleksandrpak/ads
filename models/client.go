package models

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aleksandrpak/ads/system/geoip"

	"gopkg.in/mgo.v2/bson"
)

type Client struct {
	ID            bson.ObjectId `bson:"_id,omitempty"`
	InternalID    string        `bson:"internalId,omitempty"`
	AppleID       string        `bson:"appleId,omitempty"`
	AppleIDSha1   string        `bson:"appleIdSha1,omitempty"`
	AppleIDMd5    string        `bson:"appleIdMd5,omitempty"`
	AndroidID     string        `bson:"androidId,omitempty"`
	AndroidIDSha1 string        `bson:"androidAIdSha1,omitempty"`
	AndroidIDMd5  string        `bson:"androidAIdMd5,omitempty"`
	Geo           string        `bson:"geo"`
	Ip            float64       `bson:"ip"`
	Gender        string        `bson:"gender"`
	Age           float64       `bson:"age"`
	CreatedAt     time.Time     `bson:"createdAt"`
	UpdatedAt     time.Time     `bson:"updatedAt"`
}

func GetClient(g geoip.DB, r *http.Request) *Client {
	fmt.Println(r.RemoteAddr)
	fmt.Println(g.Lookup(r.RemoteAddr))
	return &Client{Age: 20}
}
