package storage

import (
	"brok/internal/models"
	"context"
)

func (s *PqStorage) DeleteTransactionsByAssetID(ctx context.Context, assetID string) error {
	_, err := s.db.ExecContext(
		ctx,
		`delete from transactions where asset_id=$1`,
		assetID,
	)
	return err
}

func (s *PqStorage) GetTransactionsByAssetID(ctx context.Context, assetID string, transactions *[]models.Transaction) error {
	return s.db.SelectContext(ctx, transactions,
		`SELECT id, asset_id, amount, type, description, timestamp 
		FROM transactions 
		WHERE asset_id = $1`,
		assetID)
}

func (s *PqStorage) CreateTransaction(ctx context.Context, transaction models.Transaction) error {
	_, err := s.db.NamedExecContext(
		ctx,
		`INSERT INTO transactions (id, asset_id, amount, type, description, timestamp)
		VALUES (:id, :asset_id, :amount, :type, :description, :timestamp)`,
		transaction,
	)
	return err
}

func (s *PqStorage) DeleteTransaction(ctx context.Context, transactionID string) error {
	_, err := s.db.ExecContext(
		ctx,
		`DELETE FROM transactions WHERE id = $1`,
		transactionID,
	)
	return err
}

func (s *PqStorage) IsTransactionOwnedByUser(ctx context.Context, transactionID string, userID string) (bool, error) {
	var exists bool
	err := s.db.GetContext(
		ctx,
		&exists,
		`SELECT EXISTS (
			SELECT 1 FROM transactions
			WHERE id = $1 AND asset_id IN (SELECT id FROM assets WHERE user_id = $2)
		)`,
		transactionID, userID,
	)
	return exists, err
}
