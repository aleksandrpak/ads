package strategy

import (
	"github.com/aleksandrpak/ads/models"
	"github.com/aleksandrpak/ads/system/log"
)

type Strategy interface {
	NextAd(client *models.Client) (*models.Ad, log.ServerError)
}
