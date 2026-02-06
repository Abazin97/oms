package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"stock/internal/domain/models"
	"stock/internal/tx"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	ErrNoAvailableSpots = errors.New("no available spots")
)

type ParkingSpot interface {
	Get(ctx context.Context, id uuid.UUID, from time.Time, to time.Time) (bool, error)
}

type SpotReservation interface {
	Create(ctx context.Context, tx tx.Tx, r *models.Reservation) error
	Get(ctx context.Context, id uuid.UUID) (*models.Reservation, error)
	Update(ctx context.Context, tx tx.Tx, id uuid.UUID, status string) error
	DeleteExpired(ctx context.Context, tx tx.Tx, now time.Time) ([]models.Reservation, error)
}

type ParkingSpotRepository struct {
	db *sql.DB
}

type SpotReservationRepository struct {
	db *sql.DB
}

func NewParkingRepository(db *sql.DB) *ParkingSpotRepository {
	return &ParkingSpotRepository{db: db}
}

func NewSpotReservationRepository(db *sql.DB) *SpotReservationRepository {
	return &SpotReservationRepository{db: db}
}

func NewPostgresDB(url string) (*sql.DB, error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

//func (r *StockRepository) Close() error {
//	return r.db.Close()
//}

func (r *ParkingSpotRepository) Get(ctx context.Context, id uuid.UUID, from time.Time, to time.Time) (bool, error) {
	const op = "stock.repository.Get"

	row := r.db.QueryRowContext(ctx,
		"SELECT EXISTS ("+
			"SELECT 1 "+
			"FROM stock.parking_spots ps "+
			"WHERE ps.parking_lot_id = $1 "+
			"AND NOT EXISTS ("+
			"SELECT 1 "+
			"FROM stock.spot_reservations sr "+
			"WHERE sr.parking_spot_id = ps.id "+
			"AND sr.status IN ('pending', 'confirmed') "+
			"AND NOT (sr.ends_at <= $2 OR sr.starts_at >= $3)));", id, from, to)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, fmt.Errorf("%s %w", op, err)
	}

	return exists, nil
}

//func (r *ParkingPlacesRepository) Lock(ctx context.Context, tx services.Tx, id uuid.UUID) (*models.ParkingPlaces, error) {
//	const op = "stock.repository.Lock"
//
//	row := tx.QueryRowContext(ctx,
//		`SELECT id, total_spots, available_spots FROM stock.parking_lots
//				WHERE id = $1
//				FOR UPDATE `, id)
//
//	var lot models.ParkingPlaces
//	if err := row.Scan(&lot.ID, &lot.TotalSpots, &lot.AvailableSpots); err != nil {
//		if errors.Is(err, sql.ErrNoRows) {
//			return &models.ParkingPlaces{}, fmt.Errorf("%s %w", op, ErrParkingNotFound)
//		}
//		return &models.ParkingPlaces{}, fmt.Errorf("%s %w", op, err)
//	}
//
//	return &lot, nil
//}
//
//func (r *ParkingPlacesRepository) Update(ctx context.Context, tx services.Tx, lot *models.ParkingPlaces) error {
//	const op = "stock.repository.Update"
//
//	res, err := tx.ExecContext(ctx,
//		`UPDATE stock.parking_lots
//				SET name = $2, total_spots = $3, available_spots = $4
//				WHERE id = $1`, lot.ID, lot.TotalSpots, lot.AvailableSpots, lot.ID)
//
//	if err != nil {
//		return fmt.Errorf("%s %w", op, err)
//	}
//
//	rows, err := res.RowsAffected()
//	if err != nil {
//		return fmt.Errorf("%s %w", op, err)
//	}
//
//	if rows == 0 {
//		return fmt.Errorf("%s %w", op, ErrParkingNotFound)
//	}
//
//	return nil
//}
