package statistic

import (
	"net/http"
	"time"

	"git.startupteam.ru/aleksandrpak/ads/models"
	"git.startupteam.ru/aleksandrpak/ads/system/log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type StatisticsCollection interface {
	GetById(id *string) (*Statistic, log.ServerError)
	GetLast(adID *bson.ObjectId, n int) (*Statistic, log.ServerError)
	GetStatisticCount(adID *bson.ObjectId) float32
	GetStatistics(adIDs *[]models.ID) *map[bson.ObjectId]float32
	SaveStatistic(adID bson.ObjectId, appID bson.ObjectId, client *models.Client) *bson.ObjectId
	SaveNextStatistic(prev *Statistic) *bson.ObjectId
}

type statisticsCollection struct {
	*statisticCache
}

type Statistic struct {
	ID     bson.ObjectId  `bson:"_id"`
	AdID   bson.ObjectId  `bson:"adId"`
	AppID  bson.ObjectId  `bson:"appId"`
	Client *models.Client `bson:"client,omitempty"`
	At     int64          `bson:"at"`
	Prev   bson.ObjectId  `bson:"prev,omitempty"`
}

func NewStatisticsCollection(c *mgo.Collection, statisticHours int64) StatisticsCollection {
	c.EnsureIndexKey(
		"adId",
		"at",
	)

	return &statisticsCollection{new(c, statisticHours)}
}

func (c *statisticsCollection) GetById(id *string) (*Statistic, log.ServerError) {
	if !bson.IsObjectIdHex(*id) {
		return nil, log.NewError(http.StatusBadRequest, "provided value is not valid object id")
	}

	var result Statistic
	err := c.FindId(bson.ObjectIdHex(*id)).One(&result)
	if err != nil {
		return nil, log.NewInternalError(err)
	}

	return &result, nil
}

func (c *statisticsCollection) GetLast(adID *bson.ObjectId, n int) (*Statistic, log.ServerError) {
	var result Statistic
	err := c.Find(bson.M{"adId": *adID}).Sort("-at").Skip(n - 1).One(&result)
	if err != nil {
		return nil, log.NewInternalError(err)
	}

	return &result, nil
}

func (c *statisticsCollection) GetStatisticCount(adID *bson.ObjectId) float32 {
	var totalCount float32

	c.lock.RLock()

	s, ok := c.cache[*adID]
	if ok {
		s.RLock()
		totalCount = s.totalCount
		s.RUnlock()
	}

	c.lock.RUnlock()

	return totalCount
}

func (c *statisticsCollection) GetStatistics(adIDs *[]models.ID) *map[bson.ObjectId]float32 {
	statistic := make(map[bson.ObjectId]float32)

	c.lock.RLock()
	for _, adID := range *adIDs {
		s, ok := c.cache[adID.ID]
		if !ok {
			continue
		}

		s.RLock()
		statistic[adID.ID] = s.totalCount
		s.RUnlock()
	}
	c.lock.RUnlock()

	return &statistic
}

func (c *statisticsCollection) SaveStatistic(adID bson.ObjectId, appID bson.ObjectId, client *models.Client) *bson.ObjectId {
	now := time.Now().UnixNano()
	id := bson.NewObjectId()

	go c.Insert(&Statistic{
		ID:     id,
		AdID:   adID,
		AppID:  appID,
		Client: client,
		At:     now,
	})

	go c.updateStatistic(adID, now)

	return &id
}

func (c *statisticsCollection) SaveNextStatistic(prev *Statistic) *bson.ObjectId {
	now := time.Now().UnixNano()
	id := bson.NewObjectId()

	go c.Insert(&Statistic{
		ID:    id,
		AdID:  prev.AdID,
		AppID: prev.AppID,
		At:    now,
		Prev:  prev.ID,
	})

	go c.updateStatistic(prev.AdID, now)

	return &id
}
