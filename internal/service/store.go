package service

import (
	"time"

	"github.com/eviltomorrow/robber-core/pkg/mysql"
	"github.com/eviltomorrow/robber-repository/internal/model"
)

func SaveStocks(stocks []*model.Stock, timeout time.Duration) (int64, error) {
	if len(stocks) == 0 {
		return 0, nil
	}

	tx, err := mysql.DB.Begin()
	if err != nil {
		return 0, nil
	}
	affected, err := model.StockWithInsertOrUpdateMany(tx, stocks, timeout)
	if err != nil {
		tx.Rollback()
		return 0, nil
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return 0, nil
	}
	return affected, nil
}

func SaveQuotes(quotes []*model.Quote, mode string, timeout time.Duration) (int64, error) {
	if len(quotes) == 0 {
		return 0, nil
	}

	var (
		codes = make([]string, 0, len(quotes))
		date  string
		count int64
		cache = make([]*model.Quote, 0, len(quotes))
	)

	tx, err := mysql.DB.Begin()
	if err != nil {
		return 0, err
	}
	for i, quote := range quotes {
		var current = quote.Date.Format("2006-01-02")
		if date == "" {
			date = current
		}

		if date == current {
			codes = append(codes, quote.Code)
			cache = append(cache, quote)
		}
		if date != current || len(quotes)-1 == i {
			_, err = model.QuoteWithDeleteManyByCodesAndDate(tx, mode, codes, date, timeout)
			if err != nil {
				tx.Rollback()
				return 0, err
			}
			affected, err := model.QuoteWithInsertMany(tx, mode, cache, timeout)
			if err != nil {
				tx.Rollback()
				return 0, err
			}
			count += affected

			date = current
			codes = codes[:0]
			cache = cache[:0]
			codes = append(codes, quote.Code)
			cache = append(cache, quote)
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return 0, err
	}
	return count, nil
}
