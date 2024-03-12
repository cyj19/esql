package esql

import (
	"context"
	"database/sql"
)

type Tx struct {
	tx     *sql.Tx
	logger Logger
}

// Execute SQL (执行原生SQL)
func (e *Tx) Exec(query string, values ...interface{}) (sql.Result, error) {
	return e.ExecContext(context.Background(), query, values...)
}

// Execute SQL (执行原生SQL)
func (e *Tx) ExecContext(ctx context.Context, query string, values ...interface{}) (sql.Result, error) {
	stmt, err := e.tx.PrepareContext(ctx, query)
	if err != nil {
		e.logger.Output(query, err, values...)
		return nil, err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, values...)
	e.logger.Output(query, err, values...)
	return result, err
}

// To query a single piece of data, the field order of v must be consistent with the column order.
// (查询单条数据，v的字段顺序必须与columns顺序一致)
func (e *Tx) QueryRow(v interface{}, query string, values ...interface{}) error {
	return e.QueryRowContext(context.Background(), v, query, values...)
}

// To query a single piece of data, the field order of v must be consistent with the column order.
// (查询单条数据，v的字段顺序必须与columns顺序一致)
func (e *Tx) QueryRowContext(ctx context.Context, v interface{}, query string, values ...interface{}) error {
	// 追加limit 1
	query += " limit 1"
	rows, err := e.tx.QueryContext(ctx, query, values...)
	if err != nil {
		e.logger.Output(query, err, values...)
		return err
	}
	defer rows.Close()

	err = unmarshalRow(v, rows, true)
	e.logger.Output(query, err, values...)
	return err
}

// To query multiple pieces of data, the field order of v must be consistent with the column order.
// (查询多条数据，v的字段顺序必须与columns顺序一致)
func (e *Tx) QueryRows(v interface{}, query string, values ...interface{}) error {
	return e.QueryRowsContext(context.Background(), v, query, values...)
}

// To query multiple pieces of data, the field order of v must be consistent with the column order.
// (查询多条数据，v的字段顺序必须与columns顺序一致)
func (e *Tx) QueryRowsContext(ctx context.Context, v interface{}, query string, values ...interface{}) error {
	rows, err := e.tx.QueryContext(ctx, query, values...)
	if err != nil {
		e.logger.Output(query, err, values...)
		return err
	}
	defer rows.Close()

	err = unmarshalRows(v, rows, true)
	e.logger.Output(query, err, values...)
	return err
}

// Commit transaction (提交事务)
func (e *Tx) Commit() error {
	return e.tx.Commit()
}

// Rollback transaction (回滚事务)
func (e *Tx) Rollback() error {
	return e.tx.Rollback()
}
