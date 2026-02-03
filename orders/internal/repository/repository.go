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
	Get(ctx context.Context, id string) (models.Order, error)
	Update(ctx context.Context, id string, order models.Order) error
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
	const op = "orders.repository.Create"

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var id string

	err = tx.QueryRowContext(ctx, `INSERT INTO orders.orders (status, customer_id, payment_link) values ($1, $2, $3) RETURNING id`,
		o.Status, o.CustomerId, o.PaymentLink).Scan(&id)
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

func (r *postgresRepository) Get(ctx context.Context, id string) (models.Order, error) {
	const op = "orders.repository.Get"
	//log.Info("Getting order with id %s", id)

	rows, err := r.db.QueryContext(ctx,
		`SELECT o.id, o.status, o.payment_link, o.customer_id, oi.product_id, oi.name, oi.quantity, oi.price
				FROM orders.orders o 
				JOIN orders.order_items oi ON (o.id = oi.order_id)
				WHERE o.id = $1`, id,
	)
	if err != nil {
		return models.Order{}, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	order := models.Order{
		Items: make([]models.Item, 0),
	}
	var items []models.Item
	isFirstRow := true

	for rows.Next() {
		var orderedItem models.Item

		if err := rows.Scan(&order.Id, &order.Status, &order.PaymentLink, &order.CustomerId, &orderedItem.Id, &orderedItem.Name, &orderedItem.Quantity, &orderedItem.Price); err != nil {
			return models.Order{}, fmt.Errorf("%s: %w", op, err)
		}

		if isFirstRow {
			order.Items = []models.Item{}
			isFirstRow = false
		}

		items = append(items, orderedItem)
	}
	if err := rows.Err(); err != nil {
		return models.Order{}, fmt.Errorf("%s: %w", op, err)
	}
	order.Items = items

	return order, rows.Err()
}

func (r *postgresRepository) Update(ctx context.Context, id string, order models.Order) error {
	const op = "orders.repository.Update"

	_, err := r.db.ExecContext(ctx,
		`UPDATE orders.orders 
				SET status = COALESCE(NULLIF($2, ''), status), payment_link = COALESCE(NULLIF($3, ''), payment_link)
				WHERE id = $1`, id, order.Status, order.PaymentLink)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
