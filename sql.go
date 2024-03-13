package esql

import (
	"context"
	"database/sql"
)

type BaseSQL interface {
	// Execute SQL (执行原生SQL)
	Exec(query string, values ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, values ...interface{}) (sql.Result, error)
	// To query a single piece of data, the field order of v must be consistent with the column order.
	// (查询单条数据，v的字段顺序必须与columns顺序一致)
	QueryRow(v interface{}, query string, values ...interface{}) error
	QueryRowContext(ctx context.Context, v interface{}, query string, values ...interface{}) error

	// To query multiple pieces of data, the field order of v must be consistent with the column order.
	// (查询多条数据，v的字段顺序必须与columns顺序一致)
	QueryRows(v interface{}, query string, values ...interface{}) error
	QueryRowsContext(ctx context.Context, v interface{}, query string, values ...interface{}) error
}

// Execute SQL (执行原生SQL)
func (e *DB) Exec(query string, values ...interface{}) (sql.Result, error) {
	return e.ExecContext(context.Background(), query, values...)
}

// Execute SQL (执行原生SQL)
func (e *DB) ExecContext(ctx context.Context, query string, values ...interface{}) (sql.Result, error) {
	stmt, err := e.db.PrepareContext(ctx, query)
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
func (e *DB) QueryRow(v interface{}, query string, values ...interface{}) error {
	return e.QueryRowContext(context.Background(), v, query, values...)
}

// To query a single piece of data, the field order of v must be consistent with the column order.
// (查询单条数据，v的字段顺序必须与columns顺序一致)
func (e *DB) QueryRowContext(ctx context.Context, v interface{}, query string, values ...interface{}) error {
	query += " limit 1"
	rows, err := e.db.QueryContext(ctx, query, values...)
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
func (e *DB) QueryRows(v interface{}, query string, values ...interface{}) error {
	return e.QueryRowsContext(context.Background(), v, query, values...)
}

// To query multiple pieces of data, the field order of v must be consistent with the column order.
// (查询多条数据，v的字段顺序必须与columns顺序一致)
func (e *DB) QueryRowsContext(ctx context.Context, v interface{}, query string, values ...interface{}) error {
	rows, err := e.db.QueryContext(ctx, query, values...)
	if err != nil {
		e.logger.Output(query, err, values...)
		return err
	}
	defer rows.Close()

	err = unmarshalRows(v, rows, true)
	e.logger.Output(query, err, values...)
	return err
}

// Open transaction (开启事务)
func (e *DB) Begin() (*Tx, error) {
	tx, err := e.db.Begin()
	if err != nil {
		return nil, err
	}

	return &Tx{tx: tx, logger: e.logger}, nil
}

// Automate transactions (自动化事务)
func (e *DB) Transaction(fn func(tx *Tx) error) error {

	panicked := true
	tx, err := e.Begin()
	if err != nil {
		return err
	}
	defer func() {
		// 发生panic或错误则回滚
		if panicked || err != nil {
			tx.Rollback()
		}
	}()

	err = fn(tx)

	if err == nil {
		err = tx.Commit()
	}

	panicked = false
	return err
}

// Generate struct by table (通过表结构生成结构体)
func (e *DB) GenStructByTable(mode, dbName, savePath string, hasTag bool) error {
	var err error
	switch mode {
	case Mysql:
		err = genStructByMysqlTable(e, dbName, savePath, hasTag)
	case Postgres:
		err = genStructByPostgresSqlTable(e, dbName, savePath, hasTag)
	case SQLite:
		err = genStructBySQLiteTable(e, dbName, savePath, hasTag)
	default:

	}

	return err
}
