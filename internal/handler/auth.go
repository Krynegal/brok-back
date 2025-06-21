package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"brok/internal/models"
	"brok/internal/storage"
	"brok/internal/utils"
)

type AuthHandler struct {
	Storage storage.Storage
}

func NewAuthHandler(s storage.Storage) *AuthHandler {
	return &AuthHandler{
		Storage: s,
	}
}

func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
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

	user, err := h.Storage.UserByID(c, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Проверка на существующий email
	exists, err := h.Storage.IsUsersMailExist(c, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check user"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "user with this email already exists"})
		return
	}

	// Хэшируем пароль
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	userID := uuid.New().String()
	createdAt := time.Now()

	newUser := &models.UserWithPassword{
		ID:           userID,
		Email:        req.Email,
		PasswordHash: string(hash),
		CreatedAt:    createdAt,
	}

	// Вставка нового пользователя
	err = h.Storage.UserCreate(c, newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// Генерация JWT
	token, err := utils.GenerateJWT(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{Token: token})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Проверка, существует ли пользователь с таким email
	user, err := h.Storage.UserByEmail(c, req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Проверка пароля (предполагаем, что пароли хэшируются)
	if !models.CheckPassword(user.PasswordHash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Генерация JWT токена
	token, err := utils.GenerateJWT(user.ID) // Генерация JWT с userID
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{Token: token})
}
