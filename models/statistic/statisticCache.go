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

	// TODO: Filter inactive
	it := c.Find(bson.M{"at": bson.M{"$gte": time.Now().Add(-time.Hour * time.Duration(statisticHours)).UnixNano()}}).Select(bson.M{"adId": 1, "at": 1}).Sort("at").Iter()
	var s Statistic
	for it.Next(&s) {
		minutes := s.At / 60000000000
		v, ok := cache.cache[s.AdID]
		if ok {
			if v.currentStatistic != nil && v.currentStatistic.unixMinutes == minutes {
				v.currentStatistic.count++
			} else {
				if v.currentStatistic != nil {
					if v.minuteStatistics == nil {
						v.minuteStatistics = &statisticHeap{}
					}

					heap.Push(v.minuteStatistics, *v.currentStatistic)
				}

				v.currentStatistic = &minuteStatistic{
					unixMinutes: minutes,
					count:       1,
				}
			}
		} else {
			v = &statistic{}
			v.currentStatistic = &minuteStatistic{
				unixMinutes: minutes,
				count:       1,
			}
			cache.cache[s.AdID] = v
		}

		v.totalCount++
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
