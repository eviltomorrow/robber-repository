package service

import (
	"log"
	"sync"
	"testing"
	"time"

	"github.com/eviltomorrow/robber-core/pkg/mysql"
	"github.com/eviltomorrow/robber-repository/internal/model"
	"github.com/stretchr/testify/assert"
)

var (
	Stock1 = &model.Stock{
		Code:            "sh600001",
		Name:            "上海银行",
		Suspend:         "正常",
		CreateTimestamp: time.Now(),
	}
	Stock2 = &model.Stock{
		Code:            "sz000001",
		Name:            "平安银行",
		Suspend:         "正常",
		CreateTimestamp: time.Now(),
	}
	Stock3 = &model.Stock{
		Code:            "sh688001",
		Name:            "隆基股份",
		Suspend:         "停牌",
		CreateTimestamp: time.Now(),
	}

	Quote1 = &model.Quote{
		Code:            "sh600001",
		Open:            10.02,
		Close:           10.33,
		High:            10.64,
		Low:             10.00,
		YesterdayClosed: 10.01,
		Volume:          123456,
		Account:         6521480,
		Date:            time.Now(),
		NumOfYear:       time.Now().YearDay(),
		Xd:              1.0,
		CreateTimestamp: time.Now(),
	}
	Quote2 = &model.Quote{
		Code:            "sz000001",
		Open:            25.32,
		Close:           24.95,
		High:            25.32,
		Low:             24.80,
		YesterdayClosed: 25.00,
		Volume:          9581412,
		Account:         9658412205,
		Date:            time.Now(),
		NumOfYear:       time.Now().YearDay(),
		Xd:              1.0,
		CreateTimestamp: time.Now(),
	}
)

var onece sync.Once

func init() {
	mysql.DSN = "root:root@tcp(127.0.0.1:3306)/robber?charset=utf8mb4&parseTime=true&loc=Local"
	mysql.MinOpen = 5
	mysql.MaxOpen = 10
	if err := mysql.Build(); err != nil {
		log.Fatal(err)
	}
	onece.Do(func() {
		truncateStock()
		truncateQuoteDay()
		truncateQuoteWeek()
	})
}

func truncateStock() {
	if _, err := mysql.DB.Exec("truncate table stock"); err != nil {
		log.Fatal(err)
	}
}

func truncateQuoteDay() {
	if _, err := mysql.DB.Exec("truncate table quote_day"); err != nil {
		log.Fatal(err)
	}
}

func truncateQuoteWeek() {
	if _, err := mysql.DB.Exec("truncate table quote_week"); err != nil {
		log.Fatal(err)
	}
}

func TestSaveStocksNormal(t *testing.T) {
	_assert := assert.New(t)
	stocks := []*model.Stock{
		Stock1,
		Stock2,
		Stock3,
	}
	affected, err := SaveStocks(stocks, timeout)
	_assert.Nil(err)
	_assert.Equal(int64(len(stocks)), affected)
}

func TestSaveStocksBlank(t *testing.T) {
	_assert := assert.New(t)
	stocks := []*model.Stock{}
	affected, err := SaveStocks(stocks, timeout)
	_assert.Nil(err)
	_assert.Equal(int64(0), affected)
}

func TestSaveStocksName(t *testing.T) {
	_assert := assert.New(t)
	oldname := Stock1.Name
	Stock1.Name = "xd上海银行"
	stocks := []*model.Stock{
		Stock1,
	}
	affected, err := SaveStocks(stocks, timeout)
	_assert.Nil(err)
	_assert.Equal(int64(1), affected)
	Stock1.Name = oldname
}

func TestSaveQuoteNormal(t *testing.T) {
	_assert := assert.New(t)
	quotes := []*model.Quote{
		Quote1,
		Quote2,
	}
	affected, err := SaveQuotes(quotes, model.Day, timeout)
	_assert.Nil(err)
	_assert.Equal(int64(len(quotes)), affected)

	affected, err = SaveQuotes(quotes, model.Week, timeout)
	_assert.Nil(err)
	_assert.Equal(int64(len(quotes)), affected)

}
