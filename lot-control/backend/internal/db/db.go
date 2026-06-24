package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func Init(dsn string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dsn), 0755); err != nil {
		return nil, fmt.Errorf("mkdir: %w", err)
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}

	db.SetMaxOpenConns(1)

	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA journal_mode = WAL`); err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	if err := seed(db); err != nil {
		return nil, fmt.Errorf("seed: %w", err)
	}

	return db, nil
}

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    username      TEXT    UNIQUE NOT NULL,
    password_hash TEXT    NOT NULL,
    full_name     TEXT    NOT NULL DEFAULT '',
    branch        TEXT    NOT NULL DEFAULT '',
    role          TEXT    NOT NULL DEFAULT 'staff',
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS skus (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    sku_code   TEXT    UNIQUE NOT NULL,
    name       TEXT    NOT NULL,
    unit       TEXT    NOT NULL DEFAULT 'cái',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS lots (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    sku_id           INTEGER NOT NULL REFERENCES skus(id) ON DELETE CASCADE,
    lot_number       TEXT    NOT NULL,
    manufacture_date TEXT    NOT NULL DEFAULT '',
    expiry_date      TEXT    NOT NULL DEFAULT '',
    qty              INTEGER NOT NULL DEFAULT 0,
    branch           TEXT    NOT NULL DEFAULT '',
    counted_by       INTEGER REFERENCES users(id),
    counted_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
    notes            TEXT    NOT NULL DEFAULT '',
    UNIQUE(sku_id, lot_number)
);
`

func migrate(db *sql.DB) error {
	_, err := db.Exec(schema)
	return err
}

func seed(db *sql.DB) error {
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE username='admin'`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		`INSERT INTO users (username, password_hash, full_name, branch, role) VALUES (?,?,?,?,?)`,
		"admin", string(hash), "Administrator", "HQ", "admin",
	)
	return err
}
