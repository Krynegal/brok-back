package storage

import (
	"context"

	"brok/internal/models"
)

func (s *PqStorage) GetUserByEmail(email string) (*models.UserWithPassword, error) {
	var user models.UserWithPassword
	err := s.db.Get(&user, `SELECT id, email, password_hash FROM users WHERE email = $1`, email)

	return &user, err
}

func (s *PqStorage) IsUsersMailExist(ctx context.Context, email string) (bool, error) {
	var exist bool
	err := s.db.Get(&exist, `SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)`, email)

	return exist, err
}

func (s *PqStorage) UserCreate(ctx context.Context, user *models.UserWithPassword) error {
	if user == nil {
		return nil
	}

	_, err := s.db.NamedExecContext(
		ctx,
		`insert into users(id, email, password_hash, base_currency, created_at)
        values (:id, :email, :password_hash, :base_currency, :created_at)
        on conflict(id) do update
            set email=excluded.email,
                password_hash=excluded.password_hash,
                base_currency=excluded.base_currency`,
		user)
	if err != nil {
		return err
	}

	return nil
}

func (s *PqStorage) UserByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	err := s.db.Get(&user, `SELECT id, email, base_currency, created_at FROM users WHERE id = $1`, userID)

	return &user, err
}

func (s *PqStorage) UserByEmail(ctx context.Context, email string) (*models.UserWithPassword, error) {
	var user models.UserWithPassword
	err := s.db.GetContext(ctx, &user, `SELECT id, email, password_hash, base_currency, created_at FROM users WHERE email = $1`, email)
	return &user, err
}

func (s *PqStorage) UserSet(ctx context.Context, user *models.User) error {
	_, err := s.db.NamedExecContext(
		ctx,
		`UPDATE users SET email = :email WHERE id = :id`,
		user,
	)
	return err
}
