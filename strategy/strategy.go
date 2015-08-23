package strategy

import "github.com/aleksandrpak/ads/models"

type Strategy interface {
	NextAd(client *models.Client) (*models.Ad, error)
}
