package database

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func (s *service) CreateTables(ctx context.Context) error {
	return s.withTransaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			display_name TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS credentials (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			public_key BLOB NOT NULL,
			credential_id BLOB NOT NULL,
			sign_count INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);`)
		if err != nil {
			return err
		}

		return nil
	})
}
