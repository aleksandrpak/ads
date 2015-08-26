package statistic

import (
	"time"

	"git.startupteam.ru/aleksandrpak/ads/models"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type StatisticsCollection interface {
	SaveStatistic(adID bson.ObjectId, appID bson.ObjectId, client *models.Client)
	GetStatistics(adIDs *[]bson.ObjectId, period time.Time) *map[bson.ObjectId]float32
}

type statisticsCollection struct {
	*statisticCache
}

type Statistic struct {
	AdID   bson.ObjectId `bson:"adId"`
	AppID  bson.ObjectId `bson:"appId"`
	Client models.Client `bson:"client"`
	At     int64         `bson:"at"`
}

func NewStatisticsCollection(c *mgo.Collection, statisticHours int64) StatisticsCollection {
	c.EnsureIndexKey(
		"adId",
		"at",
	)

	return &statisticsCollection{new(c, statisticHours)}
}

func (c *statisticsCollection) SaveStatistic(adID bson.ObjectId, appID bson.ObjectId, client *models.Client) {
	now := time.Now().UnixNano()

	go c.Insert(&Statistic{
		AdID:   adID,
		AppID:  appID,
		Client: *client,
		At:     now,
	})

	c.updateStatistic(adID, now)
}

func (c *statisticsCollection) GetStatistics(adIDs *[]bson.ObjectId, period time.Time) *map[bson.ObjectId]float32 {
	statistic := make(map[bson.ObjectId]float32)

	c.lock.RLock()
	for _, adID := range *adIDs {
		s, ok := c.cache[adID]
		if !ok {
			continue
		}

		s.RLock()
		statistic[adID] = s.totalCount
		s.RUnlock()
	}
	c.lock.RUnlock()

	return &statistic
}
