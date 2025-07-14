package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"time"

	"brok/internal/models"
	"brok/internal/storage"
)

// ExchangeRateService сервис для работы с курсами валют
type ExchangeRateService struct {
	storage storage.Storage
	apiKey  string
	apiURL  string
}

// NewExchangeRateService создает новый сервис курсов валют
func NewExchangeRateService(storage storage.Storage) *ExchangeRateService {
	return &ExchangeRateService{
		storage: storage,
		apiKey:  "demo", // Используем бесплатный API для демо
		apiURL:  "https://api.exchangerate-api.com/v4/latest/",
	}
}

// ExchangeRateAPIResponse ответ от API курсов валют
type ExchangeRateAPIResponse struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float64 `json:"rates"`
}

// ShouldUpdateRates проверяет, нужно ли обновлять курсы валют
func (s *ExchangeRateService) ShouldUpdateRates(ctx context.Context, updateInterval time.Duration) (bool, error) {
	lastUpdate, err := s.storage.GetLastExchangeRateUpdate(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get last update time: %w", err)
	}

	// Если курсов еще нет, нужно обновить
	if lastUpdate == nil {
		return true, nil
	}

	// Проверяем, прошло ли достаточно времени с последнего обновления
	timeSinceLastUpdate := time.Since(*lastUpdate)
	return timeSinceLastUpdate >= updateInterval, nil
}

// UpdateExchangeRatesIfNeeded обновляет курсы валют только если прошло достаточно времени
func (s *ExchangeRateService) UpdateExchangeRatesIfNeeded(ctx context.Context, updateInterval time.Duration) error {
	shouldUpdate, err := s.ShouldUpdateRates(ctx, updateInterval)
	if err != nil {
		return fmt.Errorf("failed to check if update is needed: %w", err)
	}

	if !shouldUpdate {
		lastUpdate, _ := s.storage.GetLastExchangeRateUpdate(ctx)
		if lastUpdate != nil {
			log.Printf("⏰ Курсы валют обновлены %v назад, пропускаем обновление", time.Since(*lastUpdate).Round(time.Minute))
		}
		return nil
	}

	log.Printf("🔄 Обновление курсов валют (последнее обновление было более %v назад)", updateInterval)
	return s.UpdateExchangeRates(ctx)
}

// UpdateExchangeRates обновляет курсы валют из API
func (s *ExchangeRateService) UpdateExchangeRates(ctx context.Context) error {
	// Получаем список поддерживаемых валют
	currencies := models.GetSupportedCurrencies()

	for _, currency := range currencies {
		if !currency.IsSupported {
			continue
		}

		// Получаем курсы для каждой валюты
		if err := s.updateRatesForCurrency(ctx, currency.Code); err != nil {
			log.Printf("Error updating rates for %s: %v", currency.Code, err)
			continue
		}
	}

	return nil
}

// updateRatesForCurrency обновляет курсы для конкретной валюты
func (s *ExchangeRateService) updateRatesForCurrency(ctx context.Context, baseCurrency string) error {
	url := s.apiURL + baseCurrency

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp ExchangeRateAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Парсим дату
	date, err := time.Parse("2006-01-02", apiResp.Date)
	if err != nil {
		return fmt.Errorf("failed to parse date: %w", err)
	}

	// Сохраняем курсы в базу
	for targetCurrency, rate := range apiResp.Rates {
		// Пропускаем не поддерживаемые валюты
		if !models.IsCurrencySupported(targetCurrency) {
			continue
		}

		// Пропускаем курс к самому себе
		if baseCurrency == targetCurrency {
			continue
		}

		exchangeRate := models.ExchangeRate{
			FromCurrency: baseCurrency,
			ToCurrency:   targetCurrency,
			Rate:         rate,
			Timestamp:    date,
		}

		if err := s.storage.SaveExchangeRate(ctx, exchangeRate); err != nil {
			log.Printf("Error saving rate %s->%s: %v", baseCurrency, targetCurrency, err)
		}
	}

	return nil
}

// GetExchangeRate получает курс валют
func (s *ExchangeRateService) GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string, date time.Time) (float64, error) {
	// Если валюты одинаковые, возвращаем 1
	if fromCurrency == toCurrency {
		return 1.0, nil
	}

	// Ищем курс в базе данных
	rate, err := s.storage.GetExchangeRate(ctx, fromCurrency, toCurrency, date)
	if err == nil {
		return rate, nil
	}

	// Если курс не найден, пробуем найти обратный курс
	reverseRate, err := s.storage.GetExchangeRate(ctx, toCurrency, fromCurrency, date)
	if err == nil {
		return 1.0 / reverseRate, nil
	}

	// Если ничего не найдено, возвращаем ошибку
	return 0, fmt.Errorf("exchange rate not found for %s->%s on %s", fromCurrency, toCurrency, date.Format("2006-01-02"))
}

// ConvertAmount конвертирует сумму из одной валюты в другую
func (s *ExchangeRateService) ConvertAmount(ctx context.Context, amount float64, fromCurrency, toCurrency string, date time.Time) (float64, error) {
	rate, err := s.GetExchangeRate(ctx, fromCurrency, toCurrency, date)
	if err != nil {
		return 0, err
	}

	return amount * rate, nil
}

// GetLatestExchangeRate получает последний доступный курс валют
func (s *ExchangeRateService) GetLatestExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (float64, error) {
	// Если валюты одинаковые, возвращаем 1
	if fromCurrency == toCurrency {
		return 1.0, nil
	}

	// Ищем последний курс в базе данных
	rate, err := s.storage.GetLatestExchangeRate(ctx, fromCurrency, toCurrency)
	if err == nil {
		return rate, nil
	}

	// Если курс не найден, пробуем найти обратный курс
	reverseRate, err := s.storage.GetLatestExchangeRate(ctx, toCurrency, fromCurrency)
	if err == nil {
		return 1.0 / reverseRate, nil
	}

	// Если ничего не найдено, возвращаем ошибку
	return 0, fmt.Errorf("exchange rate not found for %s->%s", fromCurrency, toCurrency)
}

// ConvertAmountLatest конвертирует сумму по последнему курсу
func (s *ExchangeRateService) ConvertAmountLatest(ctx context.Context, amount float64, fromCurrency, toCurrency string) (float64, error) {
	rate, err := s.GetLatestExchangeRate(ctx, fromCurrency, toCurrency)
	if err != nil {
		return 0, err
	}

	return amount * rate, nil
}

// RoundAmount округляет сумму до 2 знаков после запятой
func (s *ExchangeRateService) RoundAmount(amount float64) float64 {
	return math.Round(amount*100) / 100
}

// GetLastUpdateTime получает время последнего обновления курсов валют
func (s *ExchangeRateService) GetLastUpdateTime(ctx context.Context) (*time.Time, error) {
	return s.storage.GetLastExchangeRateUpdate(ctx)
}
