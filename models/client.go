package models

import (
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/aleksandrpak/ads/system/geoip"
	"github.com/aleksandrpak/ads/system/log"
)

type ClientIds struct {
	InternalID    string `bson:"internalId,omitempty"`
	AppleID       string `bson:"appleId,omitempty"`
	AppleIDSha1   string `bson:"appleIdSha1,omitempty"`
	AppleIDMd5    string `bson:"appleIdMd5,omitempty"`
	AndroidID     string `bson:"androidId,omitempty"`
	AndroidIDSha1 string `bson:"androidIdSha1,omitempty"`
	AndroidIDMd5  string `bson:"androidIdMd5,omitempty"`
}

type ClientInfo struct {
	Geo         geoip.Geo `bson:"geo,omitempty"`
	Ip          string    `bson:"ip,omitempty"`
	Gender      string    `bson:"gender,omitempty"`
	Age         byte      `bson:"age,omitempty"`
	OS          string    `bson:"os"`
	OSVersion   string    `bson:"osVersion"`
	DeviceModel string    `bson:"deviceModel"`
}

type Client struct {
	Ids  ClientIds   `bson:"ids,omitempty"`
	Info *ClientInfo `bson:"info"`
}

func GetClient(g geoip.DB, r *http.Request) (*Client, log.ServerError) {
	info, err := parseInfo(g, r)
	if err != nil {
		return nil, err
	}

	return &Client{
		Ids:  parseIds(r),
		Info: info,
	}, nil
}

func parseInfo(g geoip.DB, r *http.Request) (*ClientInfo, log.ServerError) {
	query := r.URL.Query()

	info, err := parseRequiredInfo(query)
	if err != nil {
		return nil, err
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	age, _ := strconv.ParseInt(query.Get("age"), 10, 8)

	info.Geo = g.Lookup(ip)
	info.Gender = query.Get("gender")
	info.Age = byte(age)

	return info, nil
}

func parseRequiredInfo(query url.Values) (*ClientInfo, log.ServerError) {
	os := query.Get("os")
	if os == "" {
		return nil, log.NewError(http.StatusBadRequest, "os is not specified")
	}

	osVersion := query.Get("osVersion")
	if osVersion == "" {
		return nil, log.NewError(http.StatusBadRequest, "os version is not specified")
	}

	deviceModel := query.Get("deviceModel")
	if deviceModel == "" {
		return nil, log.NewError(http.StatusBadRequest, "device model is not specified")
	}

	return &ClientInfo{
		OS:          os,
		OSVersion:   osVersion,
		DeviceModel: deviceModel,
	}, nil
}

func parseIds(r *http.Request) ClientIds {
	query := r.URL.Query()

	return ClientIds{
		InternalID:    query.Get("internalId"),
		AppleID:       query.Get("appleId"),
		AppleIDSha1:   query.Get("appleIdSha1"),
		AppleIDMd5:    query.Get("appleIdMd5"),
		AndroidID:     query.Get("androidId"),
		AndroidIDSha1: query.Get("androidIdSha1"),
		AndroidIDMd5:  query.Get("androidIdMd5"),
	}
}
