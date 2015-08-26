package statistic

import (
	"container/heap"
	"sync"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"git.startupteam.ru/aleksandrpak/ads/system/log"
)

type statistic struct {
	sync.RWMutex
	minuteStatistics *statisticHeap
	currentStatistic *minuteStatistic
	totalCount       float32
}

type statisticCache struct {
	*mgo.Collection
	lock           sync.RWMutex
	cache          map[bson.ObjectId]*statistic
	statisticHours int64
}

// TODO: Refactor cache creating
func new(c *mgo.Collection, statisticHours int64) *statisticCache {
	cache := statisticCache{c, sync.RWMutex{}, make(map[bson.ObjectId]*statistic), statisticHours}

	it := c.Pipe([]bson.M{
		bson.M{"$match": bson.M{"at": bson.M{"$gte": time.Now().Add(-time.Hour * time.Duration(statisticHours)).UnixNano()}}},
		bson.M{"$project": bson.M{
			"adId": 1,
			"minutes": bson.M{"$subtract": []bson.M{
				bson.M{"$divide": []interface{}{"$at", 60000000000}},
				bson.M{"$mod": []interface{}{bson.M{"$divide": []interface{}{"$at", 60000000000}}, 1}}}}}},
		bson.M{"$group": bson.M{
			"_id":   bson.M{"adId": "$adId", "minutes": "$minutes"},
			"count": bson.M{"$sum": 1}}},
		bson.M{"$sort": bson.M{"minutes": 1}},
	}).Iter()

	var s struct {
		ID struct {
			AdID    bson.ObjectId `bson:"adId"`
			Minutes int64         `bson:"minutes"`
		} `bson:"_id"`
		Count float32 `bson:"count"`
	}

	for it.Next(&s) {
		v, ok := cache.cache[s.ID.AdID]
		if !ok {
			v = &statistic{}
			v.minuteStatistics = &statisticHeap{}
			cache.cache[s.ID.AdID] = v
		}

		heap.Push(v.minuteStatistics, minuteStatistic{
			unixMinutes: s.ID.Minutes,
			count:       s.Count,
		})

		v.totalCount += s.Count
	}

	if err := it.Close(); err != nil {
		log.Fatal.Pf("failed to load statistic: %v", err)
	}

	go cache.refreshCache()

	return &cache
}

func (c *statisticCache) refreshCache() {
	<-time.After(time.Minute)

	now := time.Now()
	minMinutes := now.Add(-time.Hour*time.Duration(c.statisticHours)).UnixNano() / 60000000000
	nowMinutes := now.UnixNano() / 60000000000

	c.lock.RLock()
	for _, statistic := range c.cache {
		statistic.Lock()
		if statistic.minuteStatistics != nil && (*statistic.minuteStatistics)[0].unixMinutes < minMinutes {
			heap.Pop(statistic.minuteStatistics)
		}

		if statistic.currentStatistic != nil && statistic.currentStatistic.unixMinutes < nowMinutes {
			if statistic.minuteStatistics == nil {
				statistic.minuteStatistics = &statisticHeap{}
			}

			heap.Push(statistic.minuteStatistics, *statistic.currentStatistic)
			statistic.currentStatistic = nil
		}

		statistic.Unlock()
	}
	c.lock.RUnlock()

	// TODO: Get data from mongo when more than one instance alive
	// TODO: stop refreshing when shut down
	go c.refreshCache()
}

func (c *statisticCache) updateStatistic(adID bson.ObjectId, now int64) {
	c.lock.RLock()
	s, ok := c.cache[adID]
	c.lock.RUnlock()

	if !ok {
		c.lock.Lock()
		s, ok = c.cache[adID]
		if !ok {
			s = &statistic{}
			c.cache[adID] = s
		}
		c.lock.Unlock()
	}

	s.Lock()
	if s.currentStatistic == nil {
		s.currentStatistic = &minuteStatistic{
			unixMinutes: now / 60000000000,
			count:       1,
		}
	} else {
		s.currentStatistic.count++
	}

	s.totalCount++
	s.Unlock()
}
