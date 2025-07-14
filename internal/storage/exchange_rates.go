package storage

import (
	"context"
	"database/sql"
	"time"

	"brok/internal/models"
)

// SaveExchangeRate сохраняет курс валют в базу данных
func (s *PqStorage) SaveExchangeRate(ctx context.Context, rate models.ExchangeRate) error {
	const query = `
		INSERT INTO exchange_rates (from_currency, to_currency, rate, timestamp)
		VALUES (:from_currency, :to_currency, :rate, :timestamp)
		ON CONFLICT (from_currency, to_currency, timestamp) 
		DO UPDATE SET rate = EXCLUDED.rate, created_at = NOW()
	`

	_, err := s.db.NamedExecContext(ctx, query, rate)
	return err
}

// GetExchangeRate получает курс валют на конкретную дату
func (s *PqStorage) GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string, date time.Time) (float64, error) {
	const query = `
		SELECT rate 
		FROM exchange_rates 
		WHERE from_currency = $1 AND to_currency = $2 AND timestamp = $3
		ORDER BY created_at DESC 
		LIMIT 1
	`

	var rate float64
	err := s.db.GetContext(ctx, &rate, query, fromCurrency, toCurrency, date)
	if err != nil {
		return 0, err
	}

	return rate, nil
}

// GetLatestExchangeRate получает последний доступный курс валют
func (s *PqStorage) GetLatestExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (float64, error) {
	const query = `
		SELECT rate 
		FROM exchange_rates 
		WHERE from_currency = $1 AND to_currency = $2
		ORDER BY timestamp DESC, created_at DESC 
		LIMIT 1
	`

	var rate float64
	err := s.db.GetContext(ctx, &rate, query, fromCurrency, toCurrency)
	if err != nil {
		return 0, err
	}

	return rate, nil
}

// GetLastExchangeRateUpdate получает время последнего обновления курсов валют
func (s *PqStorage) GetLastExchangeRateUpdate(ctx context.Context) (*time.Time, error) {
	const query = `
		SELECT MAX(created_at) as last_update
		FROM exchange_rates
	`

	var lastUpdate sql.NullTime
	err := s.db.GetContext(ctx, &lastUpdate, query)
	if err != nil {
		return nil, err
	}

	// Если записей нет, возвращаем nil
	if !lastUpdate.Valid {
		return nil, nil
	}

	return &lastUpdate.Time, nil
}
