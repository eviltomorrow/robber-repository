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

func SaveQuotes(quotes []*model.Quote, mode string, date string, timeout time.Duration) (int64, error) {
	if len(quotes) == 0 {
		return 0, nil
	}

	var codes = make([]string, 0, len(quotes))
	for _, quote := range quotes {
		codes = append(codes, quote.Code)
	}

	tx, err := mysql.DB.Begin()
	if err != nil {
		return 0, err
	}
	_, err = model.QuoteWithDeleteManyByCodesAndDate(tx, mode, codes, date, timeout)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	affected, err := model.QuoteWithInsertMany(tx, mode, quotes, timeout)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return 0, err
	}
	return affected, nil
}
