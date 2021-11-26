package model

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/eviltomorrow/robber-core/pkg/mysql"
	"github.com/stretchr/testify/assert"
)

var stock1 = Stock{
	Code:            "sz000001",
	Name:            "平安银行",
	Suspend:         "正常",
	CreateTimestamp: time.Now(),
}

var stock2 = Stock{
	Code:            "sh601012",
	Name:            "隆基股份",
	Suspend:         "正常",
	CreateTimestamp: time.Now(),
}

var stock3 = Stock{
	Code:            "sz300075",
	Name:            "宁德时代",
	Suspend:         "正常",
	CreateTimestamp: time.Now(),
}

func deleteStockMany() {
	var _sql = "delete from stock where code in (?, ?, ?)"
	_, err := mysql.DB.Exec(_sql, stock1.Code, stock2.Code, stock3.Code)
	if err != nil {
		log.Fatal(err)
	}
}

func TestStockWithInsertMany(t *testing.T) {
	_assert := assert.New(t)
	deleteStockMany()

	var stocks = []*Stock{
		&stock1,
		&stock2,
		&stock3,
	}
	tx, err := mysql.DB.Begin()
	_assert.Nil(err)

	affected, err := StockWithInsertMany(tx, stocks, timeout)
	_assert.Nil(err)
	_assert.Equal(int64(len(stocks)), affected)

	tx.Commit()

	data, err := StockWithSelectMany(mysql.DB, []string{stock1.Code, stock2.Code, stock3.Code}, timeout)
	_assert.Nil(err)
	_assert.Equal(3, len(data))

	_assert.Nil(equal(&stock1, data[stock1.Code]))
	_assert.Nil(equal(&stock2, data[stock2.Code]))
	_assert.Nil(equal(&stock3, data[stock3.Code]))
}

func TestStockWithUpdateOne(t *testing.T) {
	_assert := assert.New(t)
	deleteStockMany()

	var stocks = []*Stock{
		&stock1,
	}
	tx, err := mysql.DB.Begin()
	_assert.Nil(err)

	affected, err := StockWithInsertMany(tx, stocks, timeout)
	_assert.Nil(err)
	_assert.Equal(int64(len(stocks)), affected)

	var newStock = &Stock{
		Name:    "平安银行XD",
		Suspend: "暂停",
	}

	affected, err = StockWithUpdateOne(tx, stock1.Code, newStock, timeout)
	_assert.Nil(err)
	_assert.Equal(int64(1), affected)
	tx.Commit()

	data, err := StockWithSelectMany(mysql.DB, []string{stock1.Code}, timeout)
	_assert.Nil(err)
	_assert.Equal(newStock.Name, data[stock1.Code].Name)
}

func TestStockWithSelectMany(t *testing.T) {
	_assert := assert.New(t)
	deleteStockMany()

	var stocks = []*Stock{
		&stock1,
		&stock2,
		&stock3,
	}
	tx, err := mysql.DB.Begin()
	_assert.Nil(err)

	affected, err := StockWithInsertMany(tx, stocks, timeout)
	_assert.Nil(err)
	_assert.Equal(int64(len(stocks)), affected)

	tx.Commit()

	data, err := StockWithSelectMany(mysql.DB, []string{stock1.Code, stock2.Code, stock3.Code}, timeout)
	_assert.Nil(err)
	_assert.Equal(len(stocks), len(data))

	_assert.Nil(equal(&stock1, data[stock1.Code]))
	_assert.Nil(equal(&stock2, data[stock2.Code]))
	_assert.Nil(equal(&stock3, data[stock3.Code]))
}

func TestStockWithInsertOrUpdateMany(t *testing.T) {
	_assert := assert.New(t)
	deleteStockMany()

	var stocks = []*Stock{
		&stock1,
		&stock2,
		&stock3,
	}
	tx, err := mysql.DB.Begin()
	_assert.Nil(err)

	affected, err := StockWithInsertOrUpdateMany(tx, stocks, timeout)
	_assert.Nil(err)
	_assert.Equal(int64(len(stocks)), affected)
	err = tx.Commit()
	_assert.Nil(err)

	data, err := StockWithSelectMany(mysql.DB, []string{stock1.Code, stock2.Code, stock3.Code}, timeout)
	_assert.Nil(err)
	_assert.Equal(len(stocks), len(data))

	_assert.Nil(equal(&stock1, data[stock1.Code]))
	_assert.Nil(equal(&stock2, data[stock2.Code]))
	_assert.Nil(equal(&stock3, data[stock3.Code]))

	// 2
	stocks = []*Stock{
		&stock1,
		&stock2,
	}
	affected, err = StockWithInsertOrUpdateMany(mysql.DB, stocks, timeout)
	_assert.Nil(err)
	_assert.Equal(int64(0), affected)

	// 3
	stock4 := Stock{
		Code:            "sz000001",
		Name:            "平安银行XD",
		Suspend:         "暂停",
		CreateTimestamp: time.Now(),
	}
	stocks = []*Stock{
		&stock4,
		&stock2,
	}
	affected, err = StockWithInsertOrUpdateMany(mysql.DB, stocks, timeout)
	_assert.Nil(err)
	_assert.Equal(int64(1), affected)
}

func equal(exepcted *Stock, actual *Stock) error {
	if exepcted.Code != actual.Code {
		return fmt.Errorf("Code not equal, exepcted: %s, actual: %s", exepcted.Code, actual.Code)
	}

	if exepcted.Name != actual.Name {
		return fmt.Errorf("Name not equal, exepcted: %s, actual: %s", exepcted.Name, actual.Name)
	}

	if exepcted.Suspend != actual.Suspend {
		return fmt.Errorf("Suspend not equal, exepcted: %s, actual: %s", exepcted.Suspend, actual.Suspend)
	}

	return nil
}
