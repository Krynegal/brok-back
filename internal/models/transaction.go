package models

import (
	"time"
)

// Transaction представляет транзакцию для актива
type Transaction struct {
	ID          string    `db:"id" json:"id"`
	AssetID     string    `db:"asset_id" json:"asset_id"`
	Amount      float64   `db:"amount" json:"amount"`
	Type        string    `db:"type" json:"type"` // 'deposit', 'withdrawal', 'buy', 'sell', 'revaluation', 'dividend'
	Description string    `db:"description" json:"description"`
	Timestamp   time.Time `db:"timestamp" json:"timestamp"`
}
