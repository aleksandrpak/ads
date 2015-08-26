package strategy

import "git.startupteam.ru/aleksandrpak/ads/models"

type Strategy interface {
	NextAd(client *models.Client) (*models.Ad, error)
}
