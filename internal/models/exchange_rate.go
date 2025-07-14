package models

import (
	"time"
)

// ExchangeRate представляет курс обмена валют
type ExchangeRate struct {
	ID           int       `db:"id" json:"id"`
	FromCurrency string    `db:"from_currency" json:"from_currency"`
	ToCurrency   string    `db:"to_currency" json:"to_currency"`
	Rate         float64   `db:"rate" json:"rate"`
	Timestamp    time.Time `db:"timestamp" json:"timestamp"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// ExchangeRateRequest запрос на получение курса валют
type ExchangeRateRequest struct {
	FromCurrency string `json:"from_currency" binding:"required,len=3"`
	ToCurrency   string `json:"to_currency" binding:"required,len=3"`
	Date         string `json:"date,omitempty"` // Опциональная дата в формате YYYY-MM-DD
}

// SupportedCurrency поддерживаемая валюта
type SupportedCurrency struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	IsBase      bool   `json:"is_base,omitempty"`
	IsSupported bool   `json:"is_supported"`
}

// Список поддерживаемых валют
var SupportedCurrencies = map[string]SupportedCurrency{
	"USD": {Code: "USD", Name: "US Dollar", Symbol: "$", IsSupported: true},
	"EUR": {Code: "EUR", Name: "Euro", Symbol: "€", IsSupported: true},
	"RUB": {Code: "RUB", Name: "Russian Ruble", Symbol: "₽", IsSupported: true},
	"GBP": {Code: "GBP", Name: "British Pound", Symbol: "£", IsSupported: true},
	"JPY": {Code: "JPY", Name: "Japanese Yen", Symbol: "¥", IsSupported: true},
	"CNY": {Code: "CNY", Name: "Chinese Yuan", Symbol: "¥", IsSupported: true},
	"CHF": {Code: "CHF", Name: "Swiss Franc", Symbol: "CHF", IsSupported: true},
	"CAD": {Code: "CAD", Name: "Canadian Dollar", Symbol: "C$", IsSupported: true},
	"AUD": {Code: "AUD", Name: "Australian Dollar", Symbol: "A$", IsSupported: true},
	"KRW": {Code: "KRW", Name: "South Korean Won", Symbol: "₩", IsSupported: true},
}

// IsCurrencySupported проверяет, поддерживается ли валюта
func IsCurrencySupported(currency string) bool {
	_, exists := SupportedCurrencies[currency]
	return exists
}

// GetSupportedCurrencies возвращает список поддерживаемых валют
func GetSupportedCurrencies() []SupportedCurrency {
	currencies := make([]SupportedCurrency, 0, len(SupportedCurrencies))
	for _, currency := range SupportedCurrencies {
		currencies = append(currencies, currency)
	}
	return currencies
}
