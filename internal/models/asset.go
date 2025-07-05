package models

import (
	"time"
)

// Asset представляет актив пользователя
type Asset struct {
	ID        string    `db:"id" json:"id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Name      string    `db:"name" json:"name"`
	Type      string    `db:"type" json:"type"`
	Balance   float64   `db:"balance" json:"balance"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`

	// XIRR доходность (не хранится в БД, только для ответа)
	Xirr *float64 `json:"xirr,omitempty"`

	// APY доходность (не хранится в БД, только для ответа)
	Apy *float64 `json:"apy,omitempty"`

	// APR доходность (не хранится в БД, только для ответа)
	Apr *float64 `json:"apr,omitempty"`

	Profit *float64 `json:"profit,omitempty"`
}
