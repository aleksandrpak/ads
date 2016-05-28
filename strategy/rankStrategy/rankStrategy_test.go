package rankStrategy

import (
	"testing"

	"gopkg.in/mgo.v2/bson"

	"github.com/aleksandrpak/ads/models"
	"github.com/aleksandrpak/ads/models/statistic"

	. "github.com/smartystreets/goconvey/convey"
)

type db struct {
	ads ads
}

func (d *db) Ads() models.AdsCollection {
	return &d.ads
}

func (d *db) Apps() models.AppsCollection {
	return nil
}

func (d *db) Views() statistic.StatisticsCollection {
	return nil
}

func (d *db) Clicks() statistic.StatisticsCollection {
	return nil
}

func (d *db) Conversions() statistic.StatisticsCollection {
	return nil
}

type dbConfig struct {
}

func (c *dbConfig) Hosts() string {
	return ""
}

func (c *dbConfig) Database() string {
	return ""
}

func (c *dbConfig) NewTrafficPercentage() float32 {
	return 0
}

func (c *dbConfig) StartViewsCount() int {
	return 0
}

func (c *dbConfig) StatisticHours() int64 {
	return 0
}

type ads struct {
}

func (a *ads) GetAdByID(adID *bson.ObjectId) (*models.Ad, error) {
	return nil, nil
}

func (a *ads) GetAdIDs(info *models.ClientInfo, startViewsCount int) (*[]models.ID, error) {
	return nil, nil
}

func (a *ads) GetNewAd(info *models.ClientInfo, startViewsCount int) (*models.Ad, error) {
	return &models.Ad{}, nil
}

func (a *ads) ToggleAd(adID *bson.ObjectId, value bool) {
}

func TestReturnsNewIfNoOtherFound(t *testing.T) {
	Convey("Given to return only old ads", t, func() {
		client := models.Client{
			Info: &models.ClientInfo{},
		}

		db := &db{ads{}}
		s := New(db, &dbConfig{})

		Convey("When next ad is requested", func() {
			ad, err := s.NextAd(&client)

			Convey("The returned ad should be new", func() {
				So(ad, ShouldNotEqual, nil)

				Convey("And error should be nil", func() {
					So(err, ShouldEqual, nil)
				})
			})
		})
	})
}
