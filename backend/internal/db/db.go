package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	return db, nil
}

func Ping(db *sql.DB) error {
	for i := 0; i < 3; i++ {
		if err := db.Ping(); err == nil {
			return nil
		} else if i == 2 {
			return err
		}
		time.Sleep(time.Second)
	}
	return nil
}
