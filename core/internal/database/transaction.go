package database

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func (s *service) withTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}
