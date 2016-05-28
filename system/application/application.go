package application

import (
	"github.com/aleksandrpak/ads/system/config"
	"github.com/aleksandrpak/ads/system/database"
	"github.com/aleksandrpak/ads/system/geoip"
	"github.com/aleksandrpak/ads/system/log"
)

type Application interface {
	Database() database.Database
	AppConfig() config.AppConfig
	GeoIP() geoip.DB
	Cleanup()
}

type application struct {
	config   config.AppConfig
	database database.Database
	geoip    geoip.DB
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

func (app *application) Cleanup() {
	log.Info.Pf("closing down geo ip")
	app.geoip.Close()
	log.Info.Pf("geo ip closed")

	log.Info.Pf("closing database connection")
	app.database.Close()
	log.Info.Pf("database connection closed")
}

func NewApplication(configFilename *string) Application {
	app := &application{}

	app.config = config.New(configFilename)
	app.database = database.Connect(app.config.DbConfig())
	app.geoip = geoip.New(app.config.GeoDataPath())

	return app
}
