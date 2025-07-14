package handler

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"brok/internal/models"
	"brok/internal/services"
)

// ExchangeRateHandler обработчик для работы с курсами валют
type ExchangeRateHandler struct {
	exchangeService *services.ExchangeRateService
}

// NewExchangeRateHandler создает новый обработчик курсов валют
func NewExchangeRateHandler(exchangeService *services.ExchangeRateService) *ExchangeRateHandler {
	return &ExchangeRateHandler{
		exchangeService: exchangeService,
	}
}

// GetSupportedCurrencies возвращает список поддерживаемых валют
func (h *ExchangeRateHandler) GetSupportedCurrencies(c *gin.Context) {
	currencies := models.GetSupportedCurrencies()
	c.JSON(http.StatusOK, currencies)
}

// GetExchangeRate возвращает курс валют
func (h *ExchangeRateHandler) GetExchangeRate(c *gin.Context) {
	fromCurrency := c.Query("from_currency")
	toCurrency := c.Query("to_currency")
	dateStr := c.Query("date")

	log.Println("fromCurrency:", fromCurrency, "toCurrency:", toCurrency, "dateStr:", dateStr)

	if fromCurrency == "" || toCurrency == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required parameters: from_currency, to_currency"})
		return
	}

	// Проверяем, поддерживаются ли валюты
	if !models.IsCurrencySupported(fromCurrency) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported from_currency: " + fromCurrency})
		return
	}
	if !models.IsCurrencySupported(toCurrency) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported to_currency: " + toCurrency})
		return
	}

	var rate float64
	var err error

	if dateStr != "" {
		// Парсим дату
		var date time.Time
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
			return
		}
		rate, err = h.exchangeService.GetExchangeRate(c, fromCurrency, toCurrency, date)
	} else {
		// Используем последний курс
		rate, err = h.exchangeService.GetLatestExchangeRate(c, fromCurrency, toCurrency)
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"from_currency": fromCurrency,
		"to_currency":   toCurrency,
		"rate":          h.exchangeService.RoundAmount(rate),
		"date":          dateStr,
	})
}

// UpdateExchangeRates обновляет курсы валют из API
func (h *ExchangeRateHandler) UpdateExchangeRates(c *gin.Context) {
	// Принудительное обновление (без проверки времени)
	if err := h.exchangeService.UpdateExchangeRates(c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update exchange rates: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "exchange rates updated successfully"})
}

// UpdateExchangeRatesIfNeeded обновляет курсы валют только если нужно
func (h *ExchangeRateHandler) UpdateExchangeRatesIfNeeded(c *gin.Context) {
	// Получаем интервал обновления из параметра (по умолчанию 1 час)
	intervalStr := c.DefaultQuery("interval", "1h")
	updateInterval, err := time.ParseDuration(intervalStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid interval format, use: 1h, 30m, etc."})
		return
	}

	if err := h.exchangeService.UpdateExchangeRatesIfNeeded(c, updateInterval); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update exchange rates: " + err.Error()})
		return
	}

	// Получаем время последнего обновления для ответа
	lastUpdate, _ := h.exchangeService.GetLastUpdateTime(c)

	response := gin.H{
		"message": "exchange rates check completed",
		"updated": lastUpdate != nil && time.Since(*lastUpdate) >= updateInterval,
	}

	if lastUpdate != nil {
		response["last_update"] = lastUpdate.Format(time.RFC3339)
		response["time_since_update"] = time.Since(*lastUpdate).String()
	}

	c.JSON(http.StatusOK, response)
}

// ConvertAmount конвертирует сумму из одной валюты в другую
func (h *ExchangeRateHandler) ConvertAmount(c *gin.Context) {
	fromCurrency := c.Query("from_currency")
	toCurrency := c.Query("to_currency")
	amountStr := c.Query("amount")
	dateStr := c.Query("date")

	if fromCurrency == "" || toCurrency == "" || amountStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required parameters: from_currency, to_currency, amount"})
		return
	}

	// Проверяем, поддерживаются ли валюты
	if !models.IsCurrencySupported(fromCurrency) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported from_currency: " + fromCurrency})
		return
	}
	if !models.IsCurrencySupported(toCurrency) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported to_currency: " + toCurrency})
		return
	}

	// Парсим сумму
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount"})
		return
	}

	var convertedAmount float64

	if dateStr != "" {
		// Парсим дату
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
			return
		}
		convertedAmount, err = h.exchangeService.ConvertAmount(c, amount, fromCurrency, toCurrency, date)
	} else {
		// Используем последний курс
		convertedAmount, err = h.exchangeService.ConvertAmountLatest(c, amount, fromCurrency, toCurrency)
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"from_currency":    fromCurrency,
		"to_currency":      toCurrency,
		"original_amount":  amount,
		"converted_amount": h.exchangeService.RoundAmount(convertedAmount),
		"date":             dateStr,
	})
}
