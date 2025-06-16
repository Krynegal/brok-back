package models

import (
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User Основная модель: для чтения данных из базы
type User struct {
	ID        string    `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// UserWithPassword Модель с паролем — используется только внутри приложения
type UserWithPassword struct {
	ID           string    `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// RegisterRequest Модель запроса на регистрацию
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginResponse Ответ при логине (JWT)
type LoginResponse struct {
	Token string `json:"token"`
}

// Хэширование пароля перед сохранением в базу данных
func HashPassword(password string) (string, error) {
	// Хэшируем пароль с солью (позволяет избежать использования одинаковых хэшей для одинаковых паролей)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error hashing password:", err)
		return "", err
	}
	return string(hashedPassword), nil
}

// Проверка пароля, введенного пользователем, с хранящимся хэшем
func CheckPassword(storedPasswordHash, inputPassword string) bool {
	// Сравниваем хэшированный пароль с тем, что ввел пользователь
	err := bcrypt.CompareHashAndPassword([]byte(storedPasswordHash), []byte(inputPassword))
	if err != nil {
		// Если пароли не совпадают, возвращаем false
		return false
	}
	// Если пароли совпали, возвращаем true
	return true
}
