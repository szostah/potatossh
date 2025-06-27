package database

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite"
)

type Database struct {
	conn *sql.DB
}

func Open(path string) (*Database, error) {
	db := Database{}
	var err error
	db.conn, err = sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	_, err = db.conn.ExecContext(context.Background(), SERVER_TABLE)
	if err != nil {
		return nil, err
	}
	_, err = db.conn.ExecContext(context.Background(), SETTINGS_TABLE)
	if err != nil {
		return nil, err
	}

	return &db, nil
}

func (db *Database) Close() error {
	return db.conn.Close()
}
