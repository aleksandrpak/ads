package rankStrategy

import (
	"errors"
	"math/rand"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/aleksandrpak/ads/models"
	"github.com/aleksandrpak/ads/strategy"
	"github.com/aleksandrpak/ads/system/config"
	"github.com/aleksandrpak/ads/system/database"
	"github.com/aleksandrpak/ads/system/log"
)

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

func (s *rankStrategy) NextAd(client *models.Client) (*models.Ad, error) {
	isNew := s.isNewRand.Float32() < s.dbConfig.NewTrafficPercentage()

	ad := s.getNewAd(isNew, client.Info)
	if ad != nil {
		return ad, nil
	}

	adIDs, ad, err := s.getAdIDs(isNew, client.Info)
	if ad != nil {
		return ad, nil
	} else if err != nil {
		return nil, err
	}

	viewsPerAd, conversionsPerAd := s.getStatistics(adIDs)
	rankPerAd, totalRank := calculateRanks(viewsPerAd, conversionsPerAd)
	adID := s.chooseAd(adIDs, rankPerAd, totalRank)

	ad, err = s.db.Ads().GetAdByID(adID)
	if err != nil {
		log.Error.Pf("failed to get ad: %v", err)
		return nil, errors.New("internal error while getting ad")
	}

	return ad, nil
}

func (s *rankStrategy) getNewAd(isNew bool, info *models.ClientInfo) *models.Ad {
	if !isNew {
		return nil
	}

	return s.db.Ads().GetNewAd(info, s.dbConfig.StartViewsCount())
}

func (s *rankStrategy) getAdIDs(isNew bool, info *models.ClientInfo) (*[]bson.ObjectId, *models.Ad, error) {
	ads, startViewsCount := s.db.Ads(), s.dbConfig.StartViewsCount()

	adIDs, err := ads.GetAdIDs(info, startViewsCount)
	if err != nil {
		var ad *models.Ad
		if !isNew {
			ad = ads.GetNewAd(info, startViewsCount)
		}

		if ad == nil {
			log.Error.Pf("failed to execute ad ids query: %v", err)
			return nil, nil, errors.New("no ads found")
		} else {
			return nil, ad, nil
		}
	}

	return adIDs, nil, nil
}

func (s *rankStrategy) getStatistics(adIDs *[]bson.ObjectId) (*map[bson.ObjectId]float32, *map[bson.ObjectId]float32) {
	period := time.Now().UTC().Add(-time.Duration(time.Hour) * time.Duration(s.dbConfig.StatisticHours()))
	viewsPerAd := s.db.Views().GetStatistics(adIDs, period)
	conversionsPerAd := s.db.Conversions().GetStatistics(adIDs, period)

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

func (s *rankStrategy) chooseAd(adIDs *[]bson.ObjectId, rankPerAd *map[bson.ObjectId]float32, totalRank float32) bson.ObjectId {
	var adID bson.ObjectId
	currentRank := float32(0)
	targetRank := s.rankRand.Float32()

	for _, id := range *adIDs {
		rank, ok := (*rankPerAd)[adID]
		if !ok {
			continue
		}

		currentRank += rank / totalRank
		if currentRank >= targetRank {
			break
		}

		adID = id
	}

	return adID
}
