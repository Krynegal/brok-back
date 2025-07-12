package models

import "time"

// CreateAssetRequest используется для данных при создании актива
type CreateAssetRequest struct {
	Name string `json:"name" binding:"required"`
	Type string `json:"type" binding:"required"`
}

// UpdateAssetRequest используется для данных при обновлении актива
type UpdateAssetRequest struct {
	Name    *string  `json:"name"`    // Используем указатели, чтобы проверять изменения
	Type    *string  `json:"type"`    // Если поле не передано, значит его не нужно обновлять
	Balance *float64 `json:"balance"` // Также указатель для учета изменения баланса
}

// CreateTransactionRequest используется для данных при создании транзакции
type CreateTransactionRequest struct {
	Amount      float64    `json:"amount"`
	Type        string     `json:"type" binding:"required,oneof=deposit withdrawal buy sell revaluation dividend"` // Тип операции
	Description string     `json:"description"`
	Timestamp   *time.Time `json:"timestamp,omitempty"` // Опциональное поле для указания времени транзакции
}

// LoginRequest Модель запроса на логин
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
