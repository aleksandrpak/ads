package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/golang/glog"
)

type DbConfig interface {
	Hosts() *string
	Database() *string
}

type AppConfig interface {
	Port() int
	GeoDataPath() *string
	DbConfig() DbConfig
}

type dbConfig struct {
	HostsValue    string `json:"hosts"`
	DatabaseValue string `json:"database"`
}

type appConfig struct {
	PortValue        int      `json:"port"`
	GeoDataPathValue string   `json:"geoDataPath"`
	DbConfigValue    dbConfig `json:"dbConfig"`
}

func (dc *dbConfig) Hosts() *string {
	return &dc.HostsValue
}

func (dc *dbConfig) Database() *string {
	return &dc.DatabaseValue
}

func (ac *appConfig) Port() int {
	return ac.PortValue
}

func (ac *appConfig) GeoDataPath() *string {
	return &ac.GeoDataPathValue
}

func (ac *appConfig) DbConfig() DbConfig {
	return &ac.DbConfigValue
}

func New(filename *string) AppConfig {
	appConfig := &appConfig{}

	data, err := ioutil.ReadFile(*filename)

	checkError(err)
	checkError(json.Unmarshal(data, &appConfig))

	return appConfig
}

func checkError(err error) {
	if err != nil {
		glog.Fatalf("Can't read configuration file: %v", err)
		panic(err)
	}
}
