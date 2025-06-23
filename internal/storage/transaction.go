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

func (s *PqStorage) GetTransactionsByAssetID(ctx context.Context, assetID string) ([]models.Transaction, error) {
	transactions := []models.Transaction{}

	rows, err := s.db.QueryxContext(ctx,
		`SELECT id, asset_id, amount, type, description, timestamp 
		FROM transactions 
		WHERE asset_id = $1`,
		assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var transaction models.Transaction
		if err := rows.StructScan(&transaction); err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
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
