package service

import (
	"testing"
	"time"

	"github.com/eviltomorrow/robber-core/pkg/mysql"
	"github.com/eviltomorrow/robber-repository/internal/model"
	"github.com/eviltomorrow/robber-repository/pkg/pb"
	"github.com/stretchr/testify/assert"
)

var date = time.Date(2021, time.December, 02, 12, 00, 00, 00, time.Local)

var (
	Metadata1 = &model.Quote{
		Code:            "sh601012",
		Open:            80.91,
		Close:           82.13,
		High:            82.13,
		Low:             79.80,
		YesterdayClosed: 81.00,
		Volume:          958164,
		Account:         79152346,
		Date:            date.Add(-24 * time.Hour),
		NumOfYear:       date.Add(-24 * time.Hour).YearDay(),
		Xd:              1.0,
		CreateTimestamp: time.Now(),
	}

	Metadata2 = &model.Quote{
		Code:            "sh601012",
		Open:            82.00,
		Close:           86.77,
		High:            87.01,
		Low:             81.80,
		YesterdayClosed: 82.13,
		Volume:          1032525,
		Account:         103625482,
		Date:            date,
		NumOfYear:       date.YearDay(),
		Xd:              1.0,
		CreateTimestamp: time.Now(),
	}

	pbdata = &pb.Metadata{
		Code:            "sh601012",
		Open:            85.07,
		Latest:          86.98,
		High:            87.64,
		Low:             85.00,
		YesterdayClosed: 86.77,
		Volume:          1245358,
		Account:         136587152,
		Date:            date.Add(24 * time.Hour).Format("2006-01-02"),
	}
)

func TestBuildQuoteDay(t *testing.T) {
	_assert := assert.New(t)
	affected, err := SaveQuotes([]*model.Quote{Metadata1, Metadata2}, model.Day, timeout)
	_assert.Nil(err)
	_assert.Equal(int64(2), affected)

	md3, err := BuildQuoteDay(pbdata, date.Add(24*time.Hour))
	_assert.Nil(err)
	_assert.Equal(float64(1.0), md3.Xd)

	pbdata.YesterdayClosed = 85.00
	md3, err = BuildQuoteDay(pbdata, date.Add(24*time.Hour))
	_assert.Nil(err)
	_assert.Equal(pbdata.YesterdayClosed/Metadata2.Close, md3.Xd)

}

func TestBuildQuoteWeek(t *testing.T) {
	_assert := assert.New(t)
	affected, err := SaveQuotes([]*model.Quote{Metadata1, Metadata2}, model.Day, timeout)
	_assert.Nil(err)
	_assert.Equal(int64(2), affected)

	md3, err := BuildQuoteWeek(Metadata1.Code, date.Add(24*time.Hour))
	_assert.Nil(err)
	_assert.Equal(float64(1.0), md3.Xd)
	_assert.Equal(Metadata2.Close, md3.Close)
	_assert.Equal(Metadata1.Low, md3.Low)
}

func TestBuildQuoteWeek2(t *testing.T) {
	var (
		offset int64 = 0
		limit  int64 = 30
		date         = time.Date(2021, time.December, 17, 0, 0, 0, 0, time.Local)

		count int
	)
	for {
		stocks, err := model.StockWithSelectRange(mysql.DB, offset, limit, timeout)
		if err != nil {
			t.Fatal(err)
		}
		for _, stock := range stocks {
			week, err := BuildQuoteWeek(stock.Code, date)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("week: %s", week)
			count++
		}
		if len(stocks) < int(limit) {
			break
		}
		offset += limit
	}

	t.Logf("count: %v\r\n", count)
}
