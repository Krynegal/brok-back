package storage

import (
	"context"

	"brok/internal/models"
)

func (s *PqStorage) AssetsByUserId(ctx context.Context, userID string) ([]models.Asset, error) {
	rows, err := s.db.QueryxContext(
		ctx,
		`SELECT id, user_id, name, type, balance, currency FROM assets WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rows.Close()
	}()

	assets := []models.Asset{}

	for rows.Next() {
		asset := models.Asset{}

		if err = rows.StructScan(&asset); err != nil {
			return nil, err
		}

		assets = append(assets, asset)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return assets, nil
}

func (s *PqStorage) AssetSet(ctx context.Context, asset models.Asset) error {
	const query = `
		insert into assets(id, user_id, name, type, balance, currency, created_at)
		values (:id, :user_id, :name, :type, :balance, :currency, :created_at)
        on conflict(id) do update
		set name=excluded.name,
                type=excluded.type,
		    balance=excluded.balance,
		    currency=excluded.currency
	`

	_, err := s.db.NamedExecContext(ctx, query, asset)
	if err != nil {
		return err
	}

	return nil
}

func (s *PqStorage) DeleteAsset(ctx context.Context, assetID string) error {
	_, err := s.db.ExecContext(
		ctx,
		`delete from assets where id=$1`,
		assetID,
	)
	return err
}

func (s *PqStorage) IsAssetOwnedByUser(ctx context.Context, assetID string, userID string) (bool, error) {
	var exists bool
	err := s.db.GetContext(
		ctx,
		&exists,
		`SELECT EXISTS (
		SELECT 1 FROM assets WHERE id = $1 AND user_id = $2
	)`,
		assetID, userID,
	)
	return exists, err
}

// UpdateAssetBalance обновляет баланс актива на заданную величину
func (s *PqStorage) UpdateAssetBalance(ctx context.Context, assetID string, balanceChange float64) error {
	return s.UpdateAssetBalanceTx(ctx, s.db, assetID, balanceChange)
}

// UpdateAssetBalanceTx обновляет баланс актива на заданную величину через транзакцию
func (s *PqStorage) UpdateAssetBalanceTx(ctx context.Context, tx Tx, assetID string, balanceChange float64) error {
	_, err := tx.ExecContext(
		ctx,
		`UPDATE assets SET balance = balance + $1 WHERE id = $2`,
		balanceChange, assetID,
	)
	return err
}
