package storage

import (
	"context"

	"github.com/jmoiron/sqlx"

	"brok/internal/models"
	"database/sql"
)

// Storage интерфейс хранилища
type Storage interface {
	// user
	UserByEmail(ctx context.Context, email string) (*models.UserWithPassword, error)
	IsUsersMailExist(ctx context.Context, email string) (bool, error)
	UserByID(ctx context.Context, userID string) (*models.User, error)
	UserCreate(ctx context.Context, user *models.UserWithPassword) error
	UserSet(ctx context.Context, user *models.User) error

	// asset
	AssetsByUserId(ctx context.Context, userID string) ([]models.Asset, error)
	AssetSet(ctx context.Context, asset models.Asset) error
	DeleteAsset(ctx context.Context, assetID string) error
	IsAssetOwnedByUser(ctx context.Context, assetID string, userID string) (bool, error)
	UpdateAssetBalance(ctx context.Context, assetID string, balanceChange float64) error
	UpdateAssetBalanceTx(ctx context.Context, tx Tx, assetID string, balanceChange float64) error

	// transaction
	GetTransactionsByAssetID(ctx context.Context, assetID string) ([]models.Transaction, error)
	GetTransactionsByAssetIDTx(ctx context.Context, tx Tx, assetID string) ([]models.Transaction, error)
	CreateTransaction(ctx context.Context, transaction models.Transaction) error
	DeleteTransaction(ctx context.Context, transactionID string) error
	IsTransactionOwnedByUser(ctx context.Context, transactionID string, userID string) (bool, error)
	DeleteTransactionsByAssetID(ctx context.Context, assetID string) error
	CreateTransactionTx(ctx context.Context, tx Tx, transaction models.Transaction) error
	DeleteTransactionTx(ctx context.Context, tx Tx, transactionID string) error
	GetTransactionByID(ctx context.Context, transactionID string) (*models.Transaction, error)
	GetTransactionByIDTx(ctx context.Context, tx Tx, transactionID string) (*models.Transaction, error)
	DeleteTransactionsByAssetIDTx(ctx context.Context, tx Tx, assetID string) error

	// служебные
	Transaction(ctx context.Context, f TxFunc) (err error)
	Check() (any, error)
}

// Tx интерфейс для транзакций
type Tx interface {
	Rebind(query string) string
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error)
	GetContext(ctx context.Context, dest any, query string, args ...any) error
}
