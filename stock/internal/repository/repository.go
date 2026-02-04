package repository

import (
	"context"
	"database/sql"
	"stock/internal/domain/models"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Repository interface {
	Close() error
	Get(context.Context, string) (models.Stock, error)
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

func (r *StockRepository) Close() error {
	return r.db.Close()
}

func (r *StockRepository) Get(ctx context.Context, id string) (models.Stock, error) {

}
