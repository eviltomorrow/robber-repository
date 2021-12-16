package model

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/eviltomorrow/robber-core/pkg/mysql"
	"github.com/stretchr/testify/assert"
)

var date = time.Date(2021, time.May, 12, 0, 0, 0, 0, time.Local)
var timeout = 10 * time.Second
var m1 = &Quote{
	Code:            "sz000001",
	Open:            10.32,
	Close:           11.01,
	High:            11.27,
	Low:             10.10,
	YesterdayClosed: 10.25,
	Volume:          621532245,
	Account:         6234582445,
	Date:            date,
	NumOfYear:       date.YearDay(),
	Xd:              1.0,
	CreateTimestamp: date.Add(3 * time.Hour),
}

var m2 = &Quote{
	Code:            "sh601012",
	Open:            96.13,
	Close:           99.42,
	High:            100.03,
	Low:             96.00,
	YesterdayClosed: 96.01,
	Volume:          8924526952,
	Account:         9671856955215,
	Date:            date,
	NumOfYear:       date.YearDay(),
	Xd:              0.99,
	CreateTimestamp: date.Add(3 * time.Hour),
}

var m3 = &Quote{
	Code:            "sz300059",
	Open:            34.55,
	Close:           32.17,
	High:            34.55,
	Low:             31.99,
	YesterdayClosed: 31.57,
	Volume:          1253577415,
	Account:         35169842235,
	Date:            date,
	NumOfYear:       date.YearDay(),
	Xd:              0.5,
	CreateTimestamp: date.Add(3 * time.Hour),
}

func init() {
	mysql.DSN = "root:root@tcp(127.0.0.1:3306)/robber?charset=utf8mb4&parseTime=true&loc=Local"
	mysql.Build()
}

func deleteQuote(model string) {
	var _sql = fmt.Sprintf("delete from quote_%s where `date` = ? and code in (?, ?, ?)", model)
	_, err := mysql.DB.Exec(_sql, date.Format("2006-01-02"), m1.Code, m2.Code, m3.Code)
	if err != nil {
		log.Fatal(err)
	}
}

func TestQuoteWithInsertMany(t *testing.T) {
	_assert := assert.New(t)
	deleteQuote(Day)

	var data = []*Quote{
		m1, m2, m3,
	}

	tx, err := mysql.DB.Begin()
	_assert.Nil(err)
	affected, err := QuoteWithInsertMany(tx, Day, data, timeout)
	_assert.Nil(err)
	_assert.Equal(int64(len(data)), affected)
	err = tx.Commit()
	_assert.Nil(err)
}

func TestQuoteWithDeleteManyByCodesAndDate(t *testing.T) {
	_assert := assert.New(t)
	deleteQuote(Day)

	var data = []*Quote{
		m1, m2, m3,
	}

	tx, err := mysql.DB.Begin()
	_assert.Nil(err)
	affected, err := QuoteWithInsertMany(tx, Day, data, timeout)
	_assert.Nil(err)
	_assert.Equal(int64(len(data)), affected)

	affected, err = QuoteWithDeleteManyByCodesAndDate(tx, Day, []string{m1.Code, m2.Code, m3.Code}, date.Format("2006-01-02"), timeout)
	_assert.Nil(err)
	_assert.Equal(int64(3), affected)
	err = tx.Commit()
	_assert.Nil(err)

}

func TestQuoteWithSelectBetweenByCodeAndDate1(t *testing.T) {
	_assert := assert.New(t)
	deleteQuote(Day)

	var data = []*Quote{
		m1, m2, m3,
	}

	tx, err := mysql.DB.Begin()
	_assert.Nil(err)
	affected, err := QuoteWithInsertMany(tx, Day, data, timeout)
	_assert.Nil(err)
	_assert.Equal(int64(len(data)), affected)
	err = tx.Commit()
	_assert.Nil(err)

	models, err := QuoteWithSelectBetweenByCodeAndDate(mysql.DB, Day, m1.Code, date.Format("2006-01-02"), date.Format("2006-01-02"), timeout)
	_assert.Nil(err)
	_assert.Equal(1, len(models))

	err = equalQuote(m1, models[0])
	_assert.Nil(err)
}

func TestQuoteWithSelectManyLatest1(t *testing.T) {
	_assert := assert.New(t)
	deleteQuote(Day)

	var data = []*Quote{
		m1, m2, m3,
	}

	tx, err := mysql.DB.Begin()
	_assert.Nil(err)
	affected, err := QuoteWithInsertMany(tx, Day, data, timeout)
	_assert.Nil(err)
	_assert.Equal(int64(len(data)), affected)
	err = tx.Commit()
	_assert.Nil(err)

	models, err := QuoteWithSelectManyLatest(mysql.DB, Day, m1.Code, date.Format("2006-01-02"), 1, timeout)
	_assert.Nil(err)
	_assert.Equal(1, len(models))
}

func TestQuoteWithSelectManyLatest2(t *testing.T) {
	_assert := assert.New(t)
	models, err := QuoteWithSelectManyLatest(mysql.DB, Day, "sz000001", "2021-05-14", 30, timeout)
	_assert.Nil(err)
	for _, m := range models {
		t.Logf("data: %s\r\n\r\n", m.String())
	}
}

func TestQuoteWithSelectManyLatest3(t *testing.T) {
	_assert := assert.New(t)
	models, err := QuoteWithSelectManyLatest(mysql.DB, Week, "sz000001", "2021-05-14", 30, timeout)
	_assert.Nil(err)
	for _, m := range models {
		t.Logf("data: %s\r\n\r\n", m.String())
	}
}

func TestQuoteWithSelectBetweenByCodeAndDate2(t *testing.T) {
	_assert := assert.New(t)
	models, err := QuoteWithSelectBetweenByCodeAndDate(mysql.DB, Day, "sz000001", "2021-05-10", "2021-05-14", timeout)
	_assert.Nil(err)
	for _, m := range models {
		t.Logf("data: %s\r\n\r\n", m.String())
	}
}

func equalQuote(expected, actual *Quote) error {
	if expected.Code != actual.Code {
		return fmt.Errorf("code not equal, expected: %v, actual: %v", expected.Code, actual.Code)
	}
	if expected.Open != actual.Open {
		return fmt.Errorf("open not equal, expected: %v, actual: %v", expected.Code, actual.Code)
	}
	if expected.Close != actual.Close {
		return fmt.Errorf("close not equal, expected: %v, actual: %v", expected.Close, actual.Close)
	}
	if expected.High != actual.High {
		return fmt.Errorf("high not equal, expected: %v, actual: %v", expected.High, actual.High)
	}
	if expected.Low != actual.Low {
		return fmt.Errorf("low not equal, expected: %v, actual: %v", expected.Low, actual.Low)
	}
	if expected.YesterdayClosed != actual.YesterdayClosed {
		return fmt.Errorf("yesterday_closed not equal, expected: %v, actual: %v", expected.YesterdayClosed, actual.YesterdayClosed)
	}
	if expected.Volume != actual.Volume {
		return fmt.Errorf("volume not equal, expected: %v, actual: %v", expected.Volume, actual.Volume)
	}
	if expected.Account != actual.Account {
		return fmt.Errorf("account not equal, expected: %v, actual: %v", expected.Account, actual.Account)
	}
	if expected.Date != actual.Date {
		return fmt.Errorf("date not equal, expected: %v, actual: %v", expected.Date, actual.Date)
	}
	if expected.NumOfYear != actual.NumOfYear {
		return fmt.Errorf("day_of_year not equal, expected: %v, actual: %v", expected.NumOfYear, actual.NumOfYear)
	}
	return nil
}

func BenchmarkQuoteWithSelectManyLatest(b *testing.B) {
	var (
		code = "sz002739"
		date = "2021-06-02"
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		QuoteWithSelectManyLatest(mysql.DB, Day, code, date, 30, timeout)
	}
}
