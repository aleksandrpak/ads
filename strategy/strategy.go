package strategy

import "github.com/aleksandrpak/ads/models"

type Strategy interface {
	NextAd(app *models.App, client *models.Client) (*models.Ad, error)
}
