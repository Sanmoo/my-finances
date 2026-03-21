package sqlite

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func MustOpen(path string) *DB {
	db, err := Open(path)
	if err != nil {
		panic(err)
	}
	return db
}

func (db *DB) Close() error {
	return db.DB.Close()
}
