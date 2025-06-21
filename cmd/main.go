package main

import (
	"log"
	"net/http"
	"os"
	"runtime/debug"

	"brok/db"
	"brok/internal/handler"
	"brok/internal/routes"
	"brok/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// Загружаем .env файл
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env файл не найден, продолжаем...")
	}

	// Подключение к БД
	db, err := db.Init()
	if err != nil {
		log.Fatalf("❌ Не удалось подключиться к базе данных: %v", err)
	}
	log.Println("✅ Подключение к базе данных установлено")

	storage := storage.New(db)

	// Настройка Gin
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "debug"
	}
	gin.SetMode(ginMode)

	authHandler := handler.NewAuthHandler(storage)
	assetHandler := handler.NewAssetHandler(storage)
	transactionHandler := handler.NewTransactionHandler(storage)
	r := gin.Default()

	// This route serves our static swagger.yaml file
	r.StaticFile("/swagger.yaml", "./swagger.yaml")

	// This route serves the interactive Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL("/swagger.yaml"), // Point the UI to our spec file
	))

	// Добавляем middleware для обработки паники
	r.Use(func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v\n%s", err, debug.Stack())
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	})

	// Регистрируем маршруты
	routes.RegisterRoutes(r, authHandler, assetHandler, transactionHandler)

	// Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Сервер запущен на порту %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Не удалось запустить сервер: %v", err)
	}
}
