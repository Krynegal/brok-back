package storage

import (
	"github.com/jmoiron/sqlx"
)

type PqStorage struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *PqStorage {
	return &PqStorage{db: db}
}

// Check проверяет доступность хранилища
func (s *PqStorage) Check() (any, error) {
	_, err := s.db.Exec("select 1")
	if err != nil {
		return false, err
	}

	return true, nil
}
