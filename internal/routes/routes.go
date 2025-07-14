package routes

import (
	"brok/internal/handler"
	"brok/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes инициализирует все маршруты
func RegisterRoutes(
	router *gin.Engine,
	authHandler *handler.AuthHandler,
	assetHandler *handler.AssetHandler,
	transactionHandler *handler.TransactionHandler,
	exchangeRateHandler *handler.ExchangeRateHandler,
) {
	// Healthcheck
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Защищённые маршруты
	api := router.Group("/api")
	api.Use(middleware.JWTAuth())
	{
		api.GET("/me", authHandler.GetCurrentUser)

		// Assets
		api.GET("/assets", assetHandler.GetAssets)
		api.POST("/assets", assetHandler.CreateAsset)
		api.PATCH("/assets/:id", assetHandler.UpdateAsset)
		api.DELETE("/assets/:id", assetHandler.DeleteAsset)

		// Transactions
		api.GET("/assets/:id/transactions", transactionHandler.GetTransactionsByAsset)
		api.POST("/assets/:id/transactions", transactionHandler.CreateTransaction)
		api.DELETE("/transactions/:id", transactionHandler.DeleteTransaction)

		// Exchange Rates
		api.GET("/currencies", exchangeRateHandler.GetSupportedCurrencies)
		api.GET("/exchange-rates", exchangeRateHandler.GetExchangeRate)
		api.POST("/exchange-rates/update", exchangeRateHandler.UpdateExchangeRates)
		api.POST("/exchange-rates/update-if-needed", exchangeRateHandler.UpdateExchangeRatesIfNeeded)
		api.GET("/convert", exchangeRateHandler.ConvertAmount)
	}
}
