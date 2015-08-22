package geoip

import (
	"net"

	"github.com/aleksandrpak/ads/system/log"
	"github.com/fiorix/freegeoip"
)

type Geo struct {
	Country struct {
		ISOCode string `maxminddb:"iso_code" bson:"isoCode,omitempty"`
	} `maxminddb:"country" bson:"country,omitempty"`
	Location struct {
		Latitude  float64 `maxminddb:"latitude" bson:"latitude,omitempty"`
		Longitude float64 `maxminddb:"longitude" bson:"longitude,omitempty"`
		TimeZone  string  `maxminddb:"time_zone" bson:"timeZone,omitempty"`
	} `maxminddb:"location" bson:"location,omitempty"`
}

type DB interface {
	Lookup(addr string) Geo
}

type db struct {
	d *freegeoip.DB
}

func New(geoDataPath string) DB {
	geoDb, err := freegeoip.Open(geoDataPath)
	if err != nil {
		log.Fatal.Pf("Can't read geo data file: %v", err)
		panic(err)
	}

	return &db{geoDb}
}

func (d *db) Lookup(ip string) Geo {
	var geo Geo
	err := d.d.Lookup(net.ParseIP(ip), &geo)
	if err != nil {
		log.Error.Pf("Can't find geo location: %v", err)
	}

	return geo
}
