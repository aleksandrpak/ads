package application

import (
	"github.com/aleksandrpak/ads/system/config"
	"github.com/aleksandrpak/ads/system/database"
)

type Application interface {
	Database() database.Database
	AppConfig() config.AppConfig
}

type application struct {
	config   config.AppConfig
	database database.Database
}

func (app *application) Database() database.Database {
	return app.database
}

func (app *application) AppConfig() config.AppConfig {
	return app.config
}

func NewApplication(configFilename *string) Application {
	app := &application{}

	app.config = config.NewAppConfig(configFilename)
	app.database = database.Connect(app.config.DbConfig())

	return app
}
