package main

import (
	"log"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"brok/db"
	_ "brok/docs" // This will be generated
	"brok/internal/handler"
	"brok/internal/routes"
	"brok/internal/storage"
)

// @title           Brok API
// @version         1.0
// @description     A financial tracking API
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
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

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
