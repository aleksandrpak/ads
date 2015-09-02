package rankStrategy

import (
	"math/rand"
	"net/http"
	"time"

	"gopkg.in/mgo.v2/bson"

	"git.startupteam.ru/aleksandrpak/ads/models"
	"git.startupteam.ru/aleksandrpak/ads/strategy"
	"git.startupteam.ru/aleksandrpak/ads/system/config"
	"git.startupteam.ru/aleksandrpak/ads/system/database"
	"git.startupteam.ru/aleksandrpak/ads/system/log"
)

const notFoundDesc string = "no suitable ad found"

// TODO: Remove logging from domain objects
type rankStrategy struct {
	isNewRand *rand.Rand
	rankRand  *rand.Rand
	db        database.Database
	dbConfig  config.DbConfig
}

func New(db database.Database, dbConfig config.DbConfig) strategy.Strategy {
	return &rankStrategy{
		isNewRand: rand.New(rand.NewSource(time.Now().Unix())),
		rankRand:  rand.New(rand.NewSource(time.Now().Add(time.Hour).Unix())),
		db:        db,
		dbConfig:  dbConfig,
	}
}

func (s *rankStrategy) NextAd(client *models.Client) (*models.Ad, log.ServerError) {
	isNew := s.isNewRand.Float32() < s.dbConfig.NewTrafficPercentage()

	ad := s.getNewAd(isNew, client.Info)
	if ad != nil {
		return ad, nil
	}

	adIDs, ad, err := s.getAdIDs(isNew, client.Info)
	if ad != nil {
		return ad, nil
	} else if err != nil {
		return nil, log.New(http.StatusNotFound, notFoundDesc, err)
	}

	viewsPerAd, conversionsPerAd := s.getStatistics(adIDs)
	rankPerAd, totalRank := calculateRanks(viewsPerAd, conversionsPerAd)
	adID := s.chooseAd(adIDs, rankPerAd, totalRank)

	ad, err = s.db.Ads().GetAdByID(&adID)
	if err != nil {
		return nil, log.New(http.StatusNotFound, notFoundDesc, err)
	}

	return ad, nil
}

func (s *rankStrategy) getNewAd(isNew bool, info *models.ClientInfo) *models.Ad {
	if !isNew {
		return nil
	}

	return s.db.Ads().GetNewAd(info, s.dbConfig.StartViewsCount())
}

func (s *rankStrategy) getAdIDs(isNew bool, info *models.ClientInfo) (*[]models.ID, *models.Ad, error) {
	ads, startViewsCount := s.db.Ads(), s.dbConfig.StartViewsCount()

	adIDs, err := ads.GetAdIDs(info, startViewsCount)
	if err != nil || len(*adIDs) == 0 {
		var ad *models.Ad
		if !isNew {
			ad = ads.GetNewAd(info, startViewsCount)
		}

		if err != nil {
			return nil, nil, err
		}

		return nil, ad, nil
	}

	return adIDs, nil, nil
}

func (s *rankStrategy) getStatistics(adIDs *[]models.ID) (*map[bson.ObjectId]float32, *map[bson.ObjectId]float32) {
	viewsPerAd := s.db.Views().GetStatistics(adIDs)
	conversionsPerAd := s.db.Conversions().GetStatistics(adIDs)

	return viewsPerAd, conversionsPerAd
}

func calculateRanks(viewsPerAd, conversionsPerAd *map[bson.ObjectId]float32) (*map[bson.ObjectId]float32, float32) {
	rankPerAd := make(map[bson.ObjectId]float32)
	totalRank := float32(0)
	for adID, views := range *viewsPerAd {
		conversions, ok := (*conversionsPerAd)[adID]

		var rank float32
		if ok {
			rank = conversions / views
		} else {
			continue
		}

		totalRank += rank
		rankPerAd[adID] = rank
	}

	return &rankPerAd, totalRank
}

func (s *rankStrategy) chooseAd(adIDs *[]models.ID, rankPerAd *map[bson.ObjectId]float32, totalRank float32) bson.ObjectId {
	var adID bson.ObjectId
	currentRank := float32(0)
	targetRank := s.rankRand.Float32()

	for _, id := range *adIDs {
		rank, ok := (*rankPerAd)[id.ID]
		if !ok {
			continue
		}

		adID = id.ID

		currentRank += rank / totalRank
		if currentRank >= targetRank {
			break
		}
	}

	return adID
}
