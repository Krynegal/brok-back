package storage

import (
	"brok/internal/models"
	"context"
)

func (s *PqStorage) DeleteTransactionsByAssetID(ctx context.Context, assetID string) error {
	return s.DeleteTransactionsByAssetIDTx(ctx, s.db, assetID)
}

func (s *PqStorage) DeleteTransactionsByAssetIDTx(ctx context.Context, tx Tx, assetID string) error {
	_, err := tx.ExecContext(
		ctx,
		`delete from transactions where asset_id=$1`,
		assetID,
	)
	return err
}

func (s *PqStorage) GetTransactionsByAssetID(ctx context.Context, assetID string) ([]models.Transaction, error) {
	return s.GetTransactionsByAssetIDTx(ctx, s.db, assetID)
}

func (s *PqStorage) GetTransactionsByAssetIDTx(ctx context.Context, tx Tx, assetID string) ([]models.Transaction, error) {
	transactions := []models.Transaction{}

	rows, err := tx.QueryxContext(ctx,
		`SELECT id, asset_id, amount, currency, type, description, timestamp 
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
	return s.CreateTransactionTx(ctx, s.db, transaction)
}

func (s *PqStorage) CreateTransactionTx(ctx context.Context, tx Tx, transaction models.Transaction) error {
	_, err := tx.NamedExecContext(
		ctx,
		`INSERT INTO transactions (id, asset_id, amount, currency, type, description, timestamp)
		VALUES (:id, :asset_id, :amount, :currency, :type, :description, :timestamp)`,
		transaction,
	)
	return err
}

func (s *PqStorage) DeleteTransaction(ctx context.Context, transactionID string) error {
	return s.DeleteTransactionTx(ctx, s.db, transactionID)
}

func (s *PqStorage) DeleteTransactionTx(ctx context.Context, tx Tx, transactionID string) error {
	_, err := tx.ExecContext(
		ctx,
		`DELETE FROM transactions WHERE id = $1`,
		transactionID,
	)
	return err
}

func (s *PqStorage) GetTransactionByID(ctx context.Context, transactionID string) (*models.Transaction, error) {
	return s.GetTransactionByIDTx(ctx, s.db, transactionID)
}

func (s *PqStorage) GetTransactionByIDTx(ctx context.Context, tx Tx, transactionID string) (*models.Transaction, error) {
	var transaction models.Transaction
	err := tx.GetContext(
		ctx,
		&transaction,
		`SELECT id, asset_id, amount, currency, type, description, timestamp FROM transactions WHERE id = $1`,
		transactionID,
	)
	if err != nil {
		return nil, err
	}
	return &transaction, nil
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
