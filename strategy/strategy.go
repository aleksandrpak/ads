package strategy

import (
	"git.startupteam.ru/aleksandrpak/ads/models"
	"git.startupteam.ru/aleksandrpak/ads/system/log"
)

type Strategy interface {
	NextAd(client *models.Client) (*models.Ad, log.ServerError)
}
