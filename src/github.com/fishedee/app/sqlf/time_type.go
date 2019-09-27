package sqlf

import (
	gosql "database/sql"
	"errors"
	"reflect"
	"strings"
	"time"
)

func initTimeSqlTypeOperation() {
	a := time.Time{}
	stringType := reflect.TypeOf(a)
	sqlTypeOperation := sqlTypeOperation{
		toArgs: func(v interface{}, in []interface{}, builder *strings.Builder) ([]interface{}, error) {
			builder.WriteByte('?')
			in = append(in, v)
			return in, nil
		},
		fromResult: func(v interface{}, rows *gosql.Rows) error {
			return errors.New("time.Time dos not support setValue")
		},
		column: func(builder *strings.Builder) error {
			return errors.New("time.Time dos not support column")
		},
		setValue: func(v interface{}, in []interface{}, builder *strings.Builder) ([]interface{}, error) {
			return nil, errors.New("time.Time dos not support setValue")
		},
	}
	sqlTypeOperationMap.Store(stringType, &sqlTypeOperation)
}

func initTimePtrSqlTypeOperation() {
	var a *time.Time
	timePtrType := reflect.TypeOf(a)
	sqlTypeOperation := sqlTypeOperation{
		toArgs: func(v interface{}, in []interface{}, builder *strings.Builder) ([]interface{}, error) {
			data := v.(*time.Time)
			builder.WriteByte('?')
			in = append(in, *data)
			return in, nil
		},
		fromResult: func(v interface{}, rows *gosql.Rows) error {
			if rows.Next() {
				err := rows.Scan(v)
				if err != nil {
					return err
				}
				return nil
			} else {
				return errors.New("has no result")
			}
		},
		column: func(builder *strings.Builder) error {
			return errors.New("*time.Time dos not support column")
		},
		setValue: func(v interface{}, in []interface{}, builder *strings.Builder) ([]interface{}, error) {
			return nil, errors.New("*time.Time dos not support setValue")
		},
	}
	sqlTypeOperationMap.Store(timePtrType, &sqlTypeOperation)
}

func initTimeSliceSqlTypeOperation() {
	a := []time.Time{}
	stringSliceType := reflect.TypeOf(a)
	sqlTypeOperation := sqlTypeOperation{
		toArgs: func(v interface{}, in []interface{}, builder *strings.Builder) ([]interface{}, error) {
			data := v.([]time.Time)
			builder.WriteString(getSqlComma(len(data)))
			for _, single := range data {
				in = append(in, single)
			}
			return in, nil
		},
		fromResult: func(v interface{}, rows *gosql.Rows) error {
			return errors.New("[]time.Time dos not support setValue")
		},
		column: func(builder *strings.Builder) error {
			return errors.New("[]time.Time dos not support column")
		},
		setValue: func(v interface{}, in []interface{}, builder *strings.Builder) ([]interface{}, error) {
			return nil, errors.New("[]time.Time dos not support setValue")
		},
	}
	sqlTypeOperationMap.Store(stringSliceType, &sqlTypeOperation)
}

func initTimeSlicePtrSqlTypeOperation() {
	var a *[]time.Time
	stringSlicePtrType := reflect.TypeOf(a)
	sqlTypeOperation := sqlTypeOperation{
		toArgs: func(v interface{}, in []interface{}, builder *strings.Builder) ([]interface{}, error) {
			data := *(v.(*[]time.Time))
			builder.WriteString(getSqlComma(len(data)))
			for _, single := range data {
				in = append(in, single)
			}
			return in, nil
		},
		fromResult: func(v interface{}, rows *gosql.Rows) error {
			data := v.(*[]time.Time)
			result := []time.Time{}
			var temp time.Time
			for rows.Next() {
				err := rows.Scan(&temp)
				if err != nil {
					return err
				}
				result = append(result, temp)
			}
			*data = result
			return nil
		},
		column: func(builder *strings.Builder) error {
			return errors.New("*[]time.Time dos not support column")
		},
		setValue: func(v interface{}, in []interface{}, builder *strings.Builder) ([]interface{}, error) {
			return nil, errors.New("*[]time.Time dos not support setValue")
		},
	}
	sqlTypeOperationMap.Store(stringSlicePtrType, &sqlTypeOperation)
}

func init() {
	initTimeSqlTypeOperation()
	initTimePtrSqlTypeOperation()
	initTimeSliceSqlTypeOperation()
	initTimeSlicePtrSqlTypeOperation()
}
