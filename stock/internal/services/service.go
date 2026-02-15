package services

import (
	"context"
	"errors"
	"fmt"
	log "log/slog"
	"stock/internal/domain/models"
	"stock/internal/repository"
	"stock/internal/tx"
	"time"

	"github.com/google/uuid"
)

var ErrNotEnoughSpots = errors.New("not enough spots")

type StockService interface {
	Reserve(ctx context.Context, lotID string, orderID string, from time.Time, to time.Time) (*models.Reservation, error)
	//Release(ctx context.Context, reservationID uuid.UUID) error
	//Confirm(ctx context.Context, reservationID uuid.UUID) error
	GetAvailability(ctx context.Context, lotID uuid.UUID, from time.Time, to time.Time) (bool, error)
}

type stockService struct {
	tx                  tx.TxManager
	parkingSpotRepo     repository.ParkingSpot
	spotReservationRepo repository.SpotReservation
}

func NewStockService(tx tx.TxManager, spotRepository repository.ParkingSpot, reservationRepository repository.SpotReservation) StockService {
	return &stockService{tx: tx, parkingSpotRepo: spotRepository, spotReservationRepo: reservationRepository}
}

func (s *stockService) GetAvailability(ctx context.Context, lotID uuid.UUID, from time.Time, to time.Time) (bool, error) {
	const op = "stock.services.GetAvailability"

	available, err := s.parkingSpotRepo.Get(ctx, lotID, from, to)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return available, nil
}

func (s *stockService) Reserve(ctx context.Context, lotID string, orderID string, from time.Time, to time.Time) (*models.Reservation, error) {
	const op = "stock.services.Reserve"

	lotUUID, err := uuid.Parse(lotID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid lotID: %w", op, err)
	}

	spotID, err := s.spotReservationRepo.Get(ctx, lotUUID, from, to)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	spotUUID, err := uuid.Parse(spotID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid spotID: %w", op, err)
	}

	orderUUID, err := uuid.Parse(orderID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid orderID: %w", op, err)
	}

	log.Info("spot id", spotUUID)

	var reservation *models.Reservation

	err = s.tx.WithTx(ctx, func(tx tx.Tx) error {

		reservation = &models.Reservation{
			ParkingSpotID: spotUUID,
			OrderID:       orderUUID,
			CreatedAt:     time.Now(),
			ExpiresAt:     time.Now().Add(15 * time.Minute),
			StartsAt:      from,
			EndsAt:        to,
			Status:        "pending",
		}

		err := s.spotReservationRepo.Create(ctx, tx, reservation)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("reservation: ", reservation)

	return reservation, nil
}

//func (s *stockService) Confirm(ctx context.Context, reservationID uuid.UUID) error {
//	const op = "stock.services.Confirm"
//
//	return s.tx.WithTx(ctx, func(tx Tx) error {
//		_, err := s.parkingPlacesRepo.Get(ctx, reservationID)
//		if err != nil {
//			return fmt.Errorf("%s: %w", op, err)
//		}
//		if err := s.spotReservationRepo.Update(ctx, tx, lot); err != nil {
//			return fmt.Errorf("%s: %w", op, err)
//		}
//
//		return nil
//	})
//}
//func (s *stockService) Release(ctx context.Context, reservationID uuid.UUID) error {
//	const op = "stock.services.Release"
//
//	return s.tx.WithTx(ctx, func(tx Tx) error {
//		return nil
//	})
//}
