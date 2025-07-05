package handler

import (
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/maksim77/goxirr"

	"brok/internal/models"
	"brok/internal/storage"
)

type AssetHandler struct {
	Storage storage.Storage
}

func NewAssetHandler(s storage.Storage) *AssetHandler {
	return &AssetHandler{
		Storage: s,
	}
}

func (h *AssetHandler) GetAssets(c *gin.Context) {
	// Извлекаем user_id из контекста, который был установлен в middleware
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id"})
		return
	}

	assets, err := h.Storage.AssetsByUserId(c, userIDStr)
	if err != nil {
		return
	}

	// Для каждого актива считаем XIRR и прибыль
	for i := range assets {
		transactions, err := h.Storage.GetTransactionsByAssetID(c, assets[i].ID)
		if err != nil {
			continue // если не удалось получить транзакции, пропускаем XIRR
		}
		var cashflows goxirr.Transactions
		var profit float64
		var totalDeposits float64
		var totalWithdrawals float64
		var totalDividends float64

		for _, tx := range transactions {
			amount := tx.Amount
			switch tx.Type {
			case "deposit":
				cashflows = append(cashflows, goxirr.Transaction{
					Date: tx.Timestamp,
					Cash: -amount, // вложения — отрицательный поток для XIRR
				})
				totalDeposits += amount
			case "withdrawal":
				cashflows = append(cashflows, goxirr.Transaction{
					Date: tx.Timestamp,
					Cash: amount, // выводы — положительный поток для XIRR
				})
				totalWithdrawals += amount
			case "buy":
				cashflows = append(cashflows, goxirr.Transaction{
					Date: tx.Timestamp,
					Cash: -amount, // покупки — отрицательный поток для XIRR
				})
			case "sell":
				cashflows = append(cashflows, goxirr.Transaction{
					Date: tx.Timestamp,
					Cash: amount, // продажи — положительный поток для XIRR
				})
			case "dividend":
				cashflows = append(cashflows, goxirr.Transaction{
					Date: tx.Timestamp,
					Cash: amount, // дивиденды — положительный поток для XIRR
				})
				totalDividends += amount
			case "revaluation":
				cashflows = append(cashflows, goxirr.Transaction{
					Date: tx.Timestamp,
					Cash: amount, // может быть + или -
				})
			}
		}

		// Чистая прибыль = Текущий баланс - Сумма вложений + Сумма выводов + Дивиденды
		profit = assets[i].Balance - totalDeposits + totalWithdrawals + totalDividends
		if len(cashflows) > 1 {
			xirr := goxirr.Xirr(cashflows)
			assets[i].Xirr = &xirr
			// APY = XIRR (эффективная годовая ставка)
			apy := xirr
			assets[i].Apy = &apy
			// APR = ln(1 + XIRR), если XIRR > -1, иначе nil
			if xirr > -1 {
				apr := math.Log1p(xirr)
				assets[i].Apr = &apr
			} else {
				assets[i].Apr = nil
			}
		}
		// Добавляем прибыль в отдельное поле
		assets[i].Profit = &profit
	}

	c.JSON(http.StatusOK, assets)
}

func (h *AssetHandler) UpdateAsset(c *gin.Context) {
	var req models.UpdateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	assetID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	asset := models.Asset{
		ID:     assetID,
		UserID: userID.(string),
	}

	if req.Name != nil {
		asset.Name = *req.Name
	}

	if req.Type != nil {
		asset.Type = *req.Type
	}

	if req.Balance != nil {
		asset.Balance = *req.Balance
	}

	err := h.Storage.AssetSet(c, asset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update asset"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "asset updated successfully"})
}

func (h *AssetHandler) CreateAsset(c *gin.Context) {
	// Извлекаем user_id из контекста, который был установлен в middleware
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Преобразуем userID в строку
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id"})
		return
	}

	var req models.CreateAssetRequest
	// Читаем запрос
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Генерируем новый UUID для актива
	assetID := uuid.New().String()

	// Данные для сохранения в БД
	asset := models.Asset{
		ID:        assetID,
		UserID:    userIDStr, //c.MustGet("user_id").(string), // Извлекаем user_id из контекста
		Name:      req.Name,
		Type:      req.Type,
		Balance:   0.0, // Начальный баланс
		CreatedAt: time.Now(),
	}

	// Вставляем новый актив в базу данных
	err := h.Storage.AssetSet(c, asset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create asset"})
		return
	}

	// Возвращаем успешный ответ
	c.JSON(http.StatusOK, gin.H{"message": "asset created successfully", "asset_id": assetID})
}

func (h *AssetHandler) DeleteAsset(c *gin.Context) {
	assetID := c.Param("id")

	// Получаем user_id из JWT-контекста
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id"})
		return
	}

	// Проверка, принадлежит ли актив пользователю
	existsAndOwned, err := h.Storage.IsAssetOwnedByUser(c, assetID, userIDStr)
	if err != nil || !existsAndOwned {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found or not owned by user"})
		return
	}

	err = h.Storage.DeleteTransactionsByAssetID(c, assetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete related transactions"})
		return
	}

	// Удаляем актив
	err = h.Storage.DeleteAsset(c, assetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete asset"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "asset and related transactions deleted successfully"})
}
