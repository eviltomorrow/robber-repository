package service

import (
	"errors"
	"time"

	"github.com/eviltomorrow/robber-core/pkg/mysql"
	"github.com/eviltomorrow/robber-core/pkg/zmath"
	"github.com/eviltomorrow/robber-core/pkg/ztime"
	"github.com/eviltomorrow/robber-repository/internal/model"
	"github.com/eviltomorrow/robber-repository/pkg/pb"
)

var (
	timeout   = 10 * time.Second
	ErrNoData = errors.New("no data")
)

func BuildQuoteDay(data *pb.Quote, date time.Time) (*model.Quote, error) {
	latest, err := model.QuoteWithSelectManyLatest(mysql.DB, model.Day, data.Code, data.Date, 1, timeout)
	if err != nil {
		return nil, err
	}

	var xd float64 = 1.0
	if len(latest) == 1 && latest[0].Close != data.YesterdayClosed {
		xd = data.YesterdayClosed / latest[0].Close
	}

	quote := &model.Quote{
		Code:            data.Code,
		Open:            data.Open,
		Close:           data.Close,
		High:            data.High,
		Low:             data.Low,
		YesterdayClosed: data.YesterdayClosed,
		Volume:          data.Volume,
		Account:         data.Account,
		Date:            date,
		NumOfYear:       date.YearDay(),
		Xd:              xd,
		CreateTimestamp: time.Now(),
	}
	return quote, nil
}

func BuildQuoteWeek(code string, date time.Time) (*model.Quote, error) {
	var (
		begin = date.AddDate(0, 0, -5).Format("2006-01-02")
		end   = date.Format("2006-01-02")
	)

	days, err := model.QuoteWithSelectBetweenByCodeAndDate(mysql.DB, model.Day, code, begin, end, timeout)
	if err != nil {
		return nil, err
	}

	if len(days) == 0 {
		return nil, ErrNoData
	}

	var (
		first, last = days[0], days[len(days)-1]
		highs       = make([]float64, 0, len(days))
		lows        = make([]float64, 0, len(days))
		volumes     = make([]uint64, 0, len(days))
		accounts    = make([]float64, 0, len(days))
	)

	var xd = 1.0
	for _, d := range days {
		highs = append(highs, d.High)
		lows = append(lows, d.Low)
		volumes = append(volumes, d.Volume)
		accounts = append(accounts, d.Account)
		if d.Xd != 1.0 {
			xd = d.Xd
		}
	}

	var week = &model.Quote{
		Code:            first.Code,
		Open:            first.Open,
		Close:           last.Close,
		High:            zmath.MaxFloat64(highs),
		Low:             zmath.MinFloat64(lows),
		YesterdayClosed: first.YesterdayClosed,
		Volume:          zmath.SumUint64(volumes),
		Account:         zmath.SumFloat64(accounts),
		Date:            date,
		NumOfYear:       ztime.YearWeek(date),
		Xd:              xd,
		CreateTimestamp: time.Now(),
	}
	return week, nil
}
