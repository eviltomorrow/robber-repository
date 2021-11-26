package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/eviltomorrow/robber-core/pkg/mysql"
	jsoniter "github.com/json-iterator/go"
)

const (
	TypeStock = "stock"
	TypeDay   = "day"
	TypeWeek  = "week"
)

var types = []string{
	TypeStock,
	TypeDay,
	TypeWeek,
}

func TaskRecordWithInsertMany(exec mysql.Exec, records []*TaskRecord, timeout time.Duration) (int64, error) {
	if len(records) == 0 {
		return 0, nil
	}

	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var fields = make([]string, 0, len(records))
	var args = make([]interface{}, 0, 3*len(records))
	for _, record := range records {
		fields = append(fields, "(false, 0, ?, ?, ?, now(), null)")
		args = append(args, record.Type)
		args = append(args, record.Date)
		args = append(args, record.Priority)
	}

	var _sql = fmt.Sprintf("insert into task_record (%s) values %s", strings.Join(modelFields, ","), strings.Join(fields, ","))
	result, err := exec.ExecContext(ctx, _sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func TaskRecordWithUpdateOneCompleted(exec mysql.Exec, count int64, date string, t string, timeout time.Duration) (int64, error) {
	if date == "" || t == "" {
		return 0, nil
	}

	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var _sql = "update task_record set completed = true, `count` = ?, modify_timestamp = now() where `type` = ? and `date` = ?"
	result, err := exec.ExecContext(ctx, _sql, count, t, date)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func TaskRecordWithDeleteManyByDate(exec mysql.Exec, date string, timeout time.Duration) (int64, error) {
	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var _sql = "delete from task_record where date = ?"
	result, err := exec.ExecContext(ctx, _sql, date)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func TaskRecordWithSelectManyByDate(exec mysql.Exec, date string, timeout time.Duration) ([]*TaskRecord, error) {
	if date == "" {
		return nil, nil
	}

	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var _sql = "select id, completed, count, type, DATE_FORMAT(date,'%Y-%m-%d'), priority, create_timestamp, modify_timestamp from task_record where date = ? order by priority asc"
	rows, err := exec.QueryContext(ctx, _sql, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records = make([]*TaskRecord, 0, 64)
	for rows.Next() {
		var record = TaskRecord{}
		if err := rows.Scan(
			&record.Id,
			&record.Completed,
			&record.Count,
			&record.Type,
			&record.Date,
			&record.Priority,
			&record.CreateTimestamp,
			&record.ModifyTimestamp,
		); err != nil {
			return nil, err
		}
		records = append(records, &record)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

type TaskRecord struct {
	Id              int64        `json:"id"`
	Completed       bool         `json:"completed"`
	Count           int          `json:"count"`
	Type            string       `json:"type"`
	Date            string       `json:"date"`
	Priority        int64        `json:"priority"`
	CreateTimestamp time.Time    `json:"create_timestamp"`
	ModifyTimestamp sql.NullTime `json:"modify_timestamp"`
}

const (
	FieldTaskRecordID              = "id"
	FieldTaskRecordCompleted       = "completed"
	FieldTaskRecordCount           = "`count`"
	FieldTaskRecordType            = "`type`"
	FieldTaskRecordDate            = "`date`"
	FieldTaskRecordPriority        = "priority"
	FieldTaskRecordCreateTimestamp = "create_timestamp"
	FieldTaskRecordModifyTimestamp = "modify_timestamp"
)

var modelFields = []string{
	FieldTaskRecordCompleted,
	FieldTaskRecordCount,
	FieldTaskRecordType,
	FieldTaskRecordDate,
	FieldTaskRecordPriority,
	FieldTaskRecordCreateTimestamp,
	FieldTaskRecordModifyTimestamp,
}

func (t *TaskRecord) String() string {
	buf, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(t)
	return string(buf)
}

func BuildTaskRecord(t time.Time) []*TaskRecord {
	var date = t.Format("2006-01-02")
	var isFriday = (t.Weekday() == time.Friday)

	var data = make([]*TaskRecord, 0, 16)
	for _, t := range types {
		switch t {
		case TypeStock:
			data = append(data, buildTaskRecord(date, TypeStock, 1))
		case TypeDay:
			data = append(data, buildTaskRecord(date, TypeDay, 2))
		case TypeWeek:
			if isFriday {
				data = append(data, buildTaskRecord(date, TypeWeek, 3))
			}
		default:
		}
	}
	return data
}

func buildTaskRecord(date string, t string, priority int64) *TaskRecord {
	return &TaskRecord{
		Completed: false,
		Count:     0,
		Type:      t,
		Date:      date,
		Priority:  priority,
	}
}
