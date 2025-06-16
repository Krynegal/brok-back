package storage

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// TxFunc функция, выполняемая внутри транзакции
type TxFunc func(ctx context.Context, tx Tx) error

// Transactor запускает транзакцию
type Transactor interface {
	BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
}

// Transaction выполнение функции внутри транзакции
func (s *PqStorage) Transaction(ctx context.Context, f TxFunc) error {
	return Transaction(ctx, s.db, f)
}

// Transaction выполнение функции внутри транзакции
func Transaction(ctx context.Context, db Transactor, f TxFunc) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	err = f(ctx, tx)
	if err != nil {
		_ = tx.Rollback()

		return err
	}

	return tx.Commit()
}
