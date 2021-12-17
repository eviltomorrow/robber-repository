package model

import (
	"log"
	"testing"

	"github.com/eviltomorrow/robber-core/pkg/mysql"
	"github.com/stretchr/testify/assert"
)

var t1 = &Task{
	Date:          "2021-12-21",
	Completed:     0,
	MetadataCount: 4500,
	StockCount:    15,
	DayCount:      4500,
	WeekCount:     4500,
	CallbackURL:   "http://www.baidu.com",
}

func deleteTask() {
	var _sql = "truncate table task"
	_, err := mysql.DB.Exec(_sql)
	if err != nil {
		log.Fatal(err)
	}
}

func TestTaskWithSelectOne(t *testing.T) {
	_assert := assert.New(t)

	deleteTask()

	tx, err := mysql.DB.Begin()
	if err != nil {
		t.Fatal(err)
	}

	affected, err := TaskWithInsertOne(tx, t1, timeout)
	_assert.Nil(err)
	_assert.Equal(int64(1), affected)
	err = tx.Commit()
	if err != nil {
		t.Fatal(err)
	}

	task, err := TaskWithSelectOne(mysql.DB, t1.Date, timeout)
	_assert.Nil(err)
	_assert.Equal(t1.Date, task.Date)
	_assert.Equal(t1.MetadataCount, task.MetadataCount)
	_assert.Equal(t1.StockCount, task.StockCount)
	_assert.Equal(t1.DayCount, task.DayCount)
	_assert.Equal(t1.WeekCount, task.WeekCount)
	_assert.Equal(t1.CallbackURL, task.CallbackURL)
}

func TestTaskWithUpdateOne(t *testing.T) {
	_assert := assert.New(t)

	// deleteTask()

	tx, err := mysql.DB.Begin()
	if err != nil {
		t.Fatal(err)
	}

	// affected, err := TaskWithInsertOne(tx, t1, timeout)
	// _assert.Nil(err)
	// _assert.Equal(int64(1), affected)

	t1.Completed = 1
	t1.StockCount = 20
	affected, err := TaskWithUpdateOne(tx, t1.Date, t1, timeout)
	_assert.Nil(err)
	_assert.Equal(int64(1), affected)

	err = tx.Commit()
	if err != nil {
		t.Fatal(err)
	}

	task, err := TaskWithSelectOne(mysql.DB, t1.Date, timeout)
	_assert.Nil(err)
	_assert.Equal(t1.Date, task.Date)
	_assert.Equal(t1.MetadataCount, task.MetadataCount)
	_assert.Equal(t1.StockCount, task.StockCount)
	_assert.Equal(t1.DayCount, task.DayCount)
	_assert.Equal(t1.WeekCount, task.WeekCount)
	_assert.Equal(t1.CallbackURL, task.CallbackURL)
}
