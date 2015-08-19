package application

import (
	"github.com/aleksandrpak/ads/system/config"
	"github.com/aleksandrpak/ads/system/database"
	"github.com/aleksandrpak/ads/system/geoip"
)

type Application interface {
	Database() database.Database
	AppConfig() config.AppConfig
	GeoIP() geoip.DB
}

type application struct {
	config config.AppConfig
	// TODO: Close database on exit
	database database.Database
	// TODO: Close geoip on exit
	geoip geoip.DB
}

func (app *application) Database() database.Database {
	return app.database
}

func (app *application) AppConfig() config.AppConfig {
	return app.config
}

func (app *application) GeoIP() geoip.DB {
	return app.geoip
}

func NewApplication(configFilename *string) Application {
	app := &application{}

	app.config = config.New(configFilename)
	app.database = database.Connect(app.config.DbConfig())
	app.geoip = geoip.New(app.config.GeoDataPath())

	return app
}
