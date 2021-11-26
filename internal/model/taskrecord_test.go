package model

import (
	"log"
	"testing"
	"time"

	"github.com/eviltomorrow/robber-core/pkg/mysql"
	"github.com/stretchr/testify/assert"
)

var d1 = "2021-05-11"

var record1 = TaskRecord{
	Completed:       false,
	Count:           0,
	Type:            TypeStock,
	Date:            d1,
	Priority:        1,
	CreateTimestamp: time.Now(),
}

var record2 = TaskRecord{
	Completed:       false,
	Count:           0,
	Type:            TypeDay,
	Date:            d1,
	Priority:        2,
	CreateTimestamp: time.Now(),
}

func deleteTaskRecordByDate() {
	var _sql = "delete from task_record where date = ?"
	_, err := mysql.DB.Exec(_sql, d1)
	if err != nil {
		log.Fatal(err)
	}
}

func deleteTaskRecordWithDate(date string) {
	var _sql = "delete from task_record where date = ?"
	_, err := mysql.DB.Exec(_sql, date)
	if err != nil {
		log.Fatal(err)
	}
}

func TestTaskRecordWithTaskRecordWithInsertMany(t *testing.T) {
	_assert := assert.New(t)
	deleteTaskRecordByDate()

	var records = []*TaskRecord{
		&record1,
		&record2,
	}

	tx, err := mysql.DB.Begin()
	_assert.Nil(err)

	affected, err := TaskRecordWithInsertMany(tx, records, timeout)
	_assert.Nil(err)
	_assert.Equal(int64(len(records)), affected)

	tx.Commit()

	data, err := TaskRecordWithSelectManyByDate(mysql.DB, d1, timeout)
	_assert.Nil(err)

	var count int
	for _, d := range data {
		for _, r := range records {
			if d.Completed == r.Completed && d.Count == r.Count && d.Date == r.Date && r.Type == d.Type {
				count++
			}
		}
	}
	_assert.Equal(len(records), count)
}

func TestTaskRecordWithSelectManyByDate(t *testing.T) {
	_assert := assert.New(t)
	deleteTaskRecordByDate()

	var records = []*TaskRecord{
		&record1,
		&record2,
	}

	tx, err := mysql.DB.Begin()
	_assert.Nil(err)

	affected, err := TaskRecordWithInsertMany(tx, records, 10*time.Second)
	_assert.Nil(err)
	_assert.Equal(int64(len(records)), affected)
	tx.Commit()

	data, err := TaskRecordWithSelectManyByDate(mysql.DB, d1, 10*time.Second)
	_assert.Nil(err)
	_assert.Equal(len(records), len(data))

	var count int
	for _, d := range data {
		for _, r := range records {
			if d.Completed == r.Completed && d.Count == r.Count && d.Date == r.Date && r.Type == d.Type {
				count++
			}
		}
	}
	_assert.Equal(len(records), count)
}

func TestTaskRecordWithUpdateOneCompleted(t *testing.T) {
	_assert := assert.New(t)
	deleteTaskRecordByDate()

	var records = []*TaskRecord{
		&record1,
		&record2,
	}
	tx, err := mysql.DB.Begin()
	_assert.Nil(err)

	affected, err := TaskRecordWithInsertMany(tx, records, 10*time.Second)
	_assert.Nil(err)
	_assert.Equal(int64(len(records)), affected)

	affected, err = TaskRecordWithUpdateOneCompleted(tx, 1024, d1, record1.Type, 10*time.Second)
	_assert.Nil(err)
	_assert.Equal(int64(1), affected)
	tx.Commit()

	data, err := TaskRecordWithSelectManyByDate(mysql.DB, d1, 10*time.Second)
	_assert.Nil(err)

	var flag bool
	for _, d := range data {
		if d.Type == record1.Type && d.Date == d1 {
			flag = true
		}
	}
	_assert.Equal(true, flag)
}

func TestBuildTaskRecord(t *testing.T) {
	_assert := assert.New(t)

	// not friday
	var date1 = time.Date(2021, time.May, 14, 0, 0, 0, 0, time.Local)
	models := BuildTaskRecord(date1)
	_assert.Equal(3, len(models))
	deleteTaskRecordWithDate(date1.Format("2006-01-02"))

	affected, err := TaskRecordWithInsertMany(mysql.DB, models, timeout)
	_assert.Nil(err)
	_assert.Equal(int64(len(models)), affected)

	old, err := TaskRecordWithSelectManyByDate(mysql.DB, date1.Format("2006-01-02"), timeout)
	_assert.Nil(err)
	_assert.Equal(len(models), len(old))

	// friday
	var date2 = time.Date(2021, time.May, 14, 0, 0, 0, 0, time.Local)
	models = BuildTaskRecord(date2)
	_assert.Equal(len(types), len(models))
	deleteTaskRecordWithDate(date2.Format("2006-01-02"))

	affected, err = TaskRecordWithInsertMany(mysql.DB, models, timeout)
	_assert.Nil(err)
	_assert.Equal(int64(len(models)), affected)

	old, err = TaskRecordWithSelectManyByDate(mysql.DB, date2.Format("2006-01-02"), timeout)
	_assert.Nil(err)
	_assert.Equal(len(models), len(old))

}
