package esql

import (
	"database/sql"
	"time"
)

type DB struct {
	db     *sql.DB
	logger Logger
}

// connection database (连接数据库)
/*
	db, err := esql.Open(esql.Mysql, "user:password@(ip:port)/dbName?charset=utf8")
    if err != nil {
       log.Fatal(err)
    }
*/
func Open(dialect string, dataSource string, logger Logger) (*DB, error) {
	db, err := sql.Open(dialect, dataSource)
	if err != nil {
		return nil, err
	}

	if logger == nil {
		logger = newDefaultLogger()
	}

	return &DB{db: db, logger: logger}, nil
}

func (e *DB) Ping() error {
	return e.db.Ping()
}

func (e *DB) SetMaxOpenConns(n int) {
	e.db.SetMaxOpenConns(n)
}

func (e *DB) SetMaxIdleConns(n int) {
	e.db.SetMaxIdleConns(n)
}

func (e *DB) SetConnMaxLifetime(d time.Duration) {
	e.db.SetConnMaxLifetime(d)
}

func (e *DB) SetConnMaxIdleTime(d time.Duration) {
	e.db.SetConnMaxIdleTime(d)
}

func (e *DB) DB() *sql.DB {
	return e.db
}
