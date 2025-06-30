package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"brok/internal/models"
	"brok/internal/storage"
)

type TransactionHandler struct {
	Storage storage.Storage
}

func NewTransactionHandler(s storage.Storage) *TransactionHandler {
	return &TransactionHandler{
		Storage: s,
	}
}

func (h *TransactionHandler) GetTransactionsByAsset(c *gin.Context) {
	// Извлекаем user_id из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	_, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id"})
		return
	}

	// Получаем asset_id из параметра
	assetID := c.Param("id")

	// Получаем транзакции для указанного актива
	transactions, err := h.Storage.GetTransactionsByAssetID(c, assetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve transactions"})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	// Извлекаем user_id из контекста
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

	var req models.CreateTransactionRequest
	// Чтение данных из запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем asset_id из параметра
	assetID := c.Param("id")

	// Проверяем, существует ли актив с таким ID
	exists, err := h.Storage.IsAssetOwnedByUser(c, assetID, userIDStr)
	if err != nil || !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "asset not found or doesn't belong to user"})
		return
	}

	// Генерация нового UUID для транзакции
	transactionID := uuid.New().String()

	transaction := models.Transaction{
		ID:          transactionID,
		AssetID:     assetID,
		Amount:      req.Amount,
		Type:        req.Type,
		Description: req.Description,
		Timestamp:   time.Now(),
	}

	// Выполняем операции в транзакции
	err = h.Storage.Transaction(c, func(ctx context.Context, tx storage.Tx) error {
		// Создаем транзакцию
		if err := h.Storage.CreateTransactionTx(ctx, tx, transaction); err != nil {
			return err
		}

		// Определяем изменение баланса
		var balanceChange float64
		switch req.Type {
		case "income":
			balanceChange = req.Amount
		case "expense":
			balanceChange = -req.Amount
		}

		// Обновляем баланс актива
		return h.Storage.UpdateAssetBalanceTx(ctx, tx, assetID, balanceChange)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transaction created successfully", "transaction_id": transactionID})
}

func (h *TransactionHandler) DeleteTransaction(c *gin.Context) {
	// Извлекаем user_id из контекста
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

	// Получаем transaction_id из параметра
	transactionID := c.Param("id")

	// Проверяем, существует ли транзакция и принадлежит ли она пользователю
	exists, err := h.Storage.IsTransactionOwnedByUser(c, transactionID, userIDStr)
	if err != nil || !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transaction not found or doesn't belong to user"})
		return
	}

	// Выполняем операции в транзакции
	err = h.Storage.Transaction(c, func(ctx context.Context, tx storage.Tx) error {
		// Получаем данные транзакции
		transaction, err := h.Storage.GetTransactionByIDTx(ctx, tx, transactionID)
		if err != nil {
			return err
		}

		// Удаляем транзакцию
		if err := h.Storage.DeleteTransactionTx(ctx, tx, transactionID); err != nil {
			return err
		}

		// Отменяем эффект удаленной транзакции
		var balanceChange float64
		switch transaction.Type {
		case "income":
			balanceChange = -transaction.Amount // отменяем доход
		case "expense":
			balanceChange = transaction.Amount // отменяем расход
		}

		return h.Storage.UpdateAssetBalanceTx(ctx, tx, transaction.AssetID, balanceChange)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transaction deleted successfully"})
}
