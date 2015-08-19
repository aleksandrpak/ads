package geoip

import (
	"net"

	"github.com/fiorix/freegeoip"
	"github.com/golang/glog"
)

type Result struct {
	Country struct {
		ISOCode string            `maxminddb:"iso_code"`
		Names   map[string]string `maxminddb:"names"`
	} `maxminddb:"country"`
	Location struct {
		Latitude  float64 `maxminddb:"latitude"`
		Longitude float64 `maxminddb:"longitude"`
		TimeZone  string  `maxminddb:"time_zone"`
	} `maxminddb:"location"`
}

type DB interface {
	Lookup(addr string) *Result
}

type db struct {
	d *freegeoip.DB
}

func New(geoDataPath *string) DB {
	geoDb, err := freegeoip.Open(*geoDataPath)
	if err != nil {
		glog.Fatalf("Can't read geo data file: %v", err)
		panic(err)
	}

	return &db{geoDb}
}

func (d *db) Lookup(addr string) *Result {
	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		glog.Errorf("Can't parse address \"%v\": %v", addr, err)
		return nil
	}

	var result Result
	err = d.d.Lookup(net.ParseIP(ip), &result)
	if err != nil {
		glog.Errorf("Can't find geo location: %v", err)
		return nil
	}

	return &result
}
