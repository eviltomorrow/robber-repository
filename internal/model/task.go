package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/eviltomorrow/robber-core/pkg/mysql"
	jsoniter "github.com/json-iterator/go"
)

func TaskWithSelectOne(exec mysql.Exec, date string, timeout time.Duration) (*Task, error) {
	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var _sql = `select date, completed, metadata_count, stock_count, day_count, week_count, callback_url, create_timestamp, modify_timestamp from task where date = ?`
	row := exec.QueryRowContext(ctx, _sql, date)
	if row.Err() != nil {
		return nil, row.Err()
	}

	var task = &Task{}
	if err := row.Scan(
		&task.Date,
		&task.Completed,
		&task.MetadataCount,
		&task.StockCount,
		&task.DayCount,
		&task.WeekCount,
		&task.CallbackURL,
		&task.CreateTimestamp,
		&task.ModifyTimestamp,
	); err != nil {
		return nil, err
	}
	return task, nil
}

func TaskWithInsertOne(exec mysql.Exec, task *Task, timeout time.Duration) (int64, error) {
	if task == nil {
		return 0, nil
	}

	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var _sql = `insert into task(date, completed, metadata_count, stock_count, day_count, week_count, callback_url, create_timestamp) values (?, ?, ?, ?, ?, ?, ?, now())`
	result, err := exec.ExecContext(ctx, _sql, task.Date, task.Completed, task.MetadataCount, task.StockCount, task.DayCount, task.WeekCount, task.CallbackURL)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func TaskWithUpdateOne(exec mysql.Exec, date string, task *Task, timeout time.Duration) (int64, error) {
	if task == nil {
		return 0, nil
	}
	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var _sql = `update task set completed = ?, metadata_count = ?, stock_count = ?, day_count = ?, week_count = ?, callback_url = ?, modify_timestamp = now() where date = ?`
	result, err := exec.ExecContext(ctx, _sql, task.Completed, task.MetadataCount, task.StockCount, task.DayCount, task.WeekCount, task.CallbackURL, date)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const (
	FieldTaskDate            = "date"
	FieldTaskCompleted       = "completed"
	FieldTaskMetadataCount   = "metadata_count"
	FieldTaskStockCount      = "stock_count"
	FieldTaskDayCount        = "day_count"
	FieldTaskWeekCount       = "week_count"
	FieldTaskCallBackURL     = "callback_url"
	FieldTaskCreateTimestamp = "create_timestamp"
	FieldTaskModifyTimestamp = "modify_timestamp"
)

var TaskFields = []string{
	FieldTaskDate,
	FieldTaskCompleted,
	FieldTaskMetadataCount,
	FieldTaskStockCount,
	FieldTaskDayCount,
	FieldTaskWeekCount,
	FieldTaskCallBackURL,
	FieldTaskCreateTimestamp,
	FieldTaskModifyTimestamp,
}

// Task
type Task struct {
	Date            string       `json:"date"`
	Completed       int8         `json:"completed"`
	MetadataCount   int64        `json:"metadata_count"`
	StockCount      int64        `json:"stock_count"`
	DayCount        int64        `json:"day_count"`
	WeekCount       int64        `json:"week_count"`
	CallbackURL     string       `json:"callback_url"`
	CreateTimestamp time.Time    `json:"create_timestamp"`
	ModifyTimestamp sql.NullTime `json:"modify_timestamp"`
}

func (t *Task) String() string {
	buf, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(t)
	return string(buf)
}
