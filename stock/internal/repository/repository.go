package repository

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Repository interface {
}

type StockRepository struct {
	db *sql.DB
}

func NewRepository(url string) (Repository, error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &StockRepository{db: db}, nil
}
