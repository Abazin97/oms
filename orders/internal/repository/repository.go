package repository

import (
	"context"
	"database/sql"
	"fmt"
	"orders/internal/domain/models"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Repository interface {
	Create(ctx context.Context, o models.Order) (string, error)
	Get(ctx context.Context, id int32) (models.Order, error)
	Update(ctx context.Context, order models.Order) error
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(url string) (Repository, error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &postgresRepository{db: db}, nil
}

func (r *postgresRepository) Close() {
	r.db.Close()
}

func (r *postgresRepository) Ping() error {
	return r.db.Ping()
}

func (r *postgresRepository) Create(ctx context.Context, o models.Order) (string, error) {
	const op = "repository.Create"

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var id string

	err = tx.QueryRowContext(ctx, `INSERT INTO orders.orders (status, payment_link) values ($1, $2) RETURNING id`,
		o.Status, o.PaymentLink).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO orders.order_items (order_id, product_id, name, quantity, price) values ($1, $2, $3, $4, $5)`)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	for _, item := range o.Items {
		if _, err = stmt.ExecContext(ctx, id, item.Id, item.Name, item.Quantity, item.Price); err != nil {
			return "", fmt.Errorf("%s: %w", op, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (r *postgresRepository) Get(ctx context.Context, id int32) (models.Order, error) {
	const op = "repository.Get"

	var order models.Order

	row := r.db.QueryRowContext(ctx, `SELECT id, status, payment_link FROM orders WHERE id = $1`, id)

	err := row.Scan(&order.Id, &order.Status, &order.PaymentLink)
	if err != nil {
		return order, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := r.db.QueryContext(ctx, `SELECT id, name, quantity, price FROM order_items WHERE order_id = $1`, order.Id)
	if err != nil {
		return order, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.Id, &item.Name, &item.Quantity, &item.Price); err != nil {
			return order, fmt.Errorf("%s: %w", op, err)
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}

func (r *postgresRepository) Update(ctx context.Context, order models.Order) error {
	const op = "repository.Update"

	return nil
}
