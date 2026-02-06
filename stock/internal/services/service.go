package services

import (
	"context"
	"errors"
	"fmt"
	"stock/internal/domain/models"
	"stock/internal/repository"
	"stock/internal/tx"
	"time"

	"github.com/google/uuid"
)

var ErrNotEnoughSpots = errors.New("not enough spots")

type StockService interface {
	Reserve(ctx context.Context, lotID uuid.UUID, orderID uuid.UUID, count int) (*models.Reservation, error)
	Release(ctx context.Context, reservationID uuid.UUID) error
	Confirm(ctx context.Context, reservationID uuid.UUID) error
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

//func (s *stockService) Confirm(ctx context.Context, reservationID uuid.UUID) error {
//	const op = "stock.services.Confirm"
//
//	return s.tx.WithTx(ctx, func(tx Tx) error {
//		_, err := s.parkingPlacesRepo.Get(ctx, reservationID)
//		if err != nil {
//			return fmt.Errorf("%s: %w", op, err)
//		}
//
//		return nil
//	})
//}
//
//func (s *stockService) Reserve(ctx context.Context, lotID uuid.UUID, orderID uuid.UUID, count int) (*models.Reservation, error) {
//	const op = "stock.services.Reserve"
//
//	var reservation *models.Reservation
//
//	err := s.tx.WithTx(ctx, func(tx Tx) error {
//		lot, err := s.parkingPlacesRepo.Lock(ctx, tx, lotID)
//		if err != nil {
//			return fmt.Errorf("%s: %w", op, err)
//		}
//
//		if lot.AvailableSpots < count {
//			return ErrNotEnoughSpots
//		}
//
//		lot.AvailableSpots -= count
//
//		if err := s.parkingPlacesRepo.Update(ctx, tx, lot); err != nil {
//			return fmt.Errorf("%s: %w", op, err)
//		}
//
//		reservation = &models.Reservation{
//			ID:           uuid.New(),
//			ParkingLotID: lotID,
//			OrderID:      orderID,
//			SpotsCount:   count,
//			ExpiresAt:    time.Now().Add(15 * time.Minute),
//		}
//
//		return nil
//	})
//
//	if err != nil {
//		return nil, fmt.Errorf("%s: %w", op, err)
//	}
//
//	return reservation, nil
//}
//
//func (s *stockService) Release(ctx context.Context, reservationID uuid.UUID) error {
//	const op = "stock.services.Release"
//
//	return s.tx.WithTx(ctx, func(tx Tx) error {
//		return nil
//	})
//}
