package db

import (
	"fmt"
	"os"

	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
)

func Init() (*sqlx.DB, error) {
	dsn := os.Getenv("DATABASE_URL")

	fmt.Println(dsn)

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
