package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gateway/rabbitmq"
	"log"
	"stock/internal/domain/models"
	"stock/internal/events"
	"stock/internal/repository"
	"stock/internal/tx"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

var ErrNotEnoughSpots = errors.New("not enough spots")

type StockService interface {
	Reserve(ctx context.Context, lotID string, orderID string, from time.Time, to time.Time) (*models.Reservation, error)
	ChangeStatus(ctx context.Context, reservationID uuid.UUID, status string) error
	GetAvailability(ctx context.Context, lotID uuid.UUID, from time.Time, to time.Time) (bool, error)
}

type stockService struct {
	tx                  tx.TxManager
	parkingSpotRepo     repository.ParkingSpot
	spotReservationRepo repository.SpotReservation
	channel             *amqp.Channel
}

func NewStockService(tx tx.TxManager, spotRepository repository.ParkingSpot, reservationRepository repository.SpotReservation, channel *amqp.Channel) StockService {
	return &stockService{tx: tx, parkingSpotRepo: spotRepository, spotReservationRepo: reservationRepository, channel: channel}
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

	event := events.StockReservedEvent{
		ReservationID: spotID,
		OrderID:       orderID,
		Status:        reservation.Status,
		StartsAt:      reservation.StartsAt,
		EndsAt:        reservation.EndsAt,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := s.channel.PublishWithContext(ctx, rabbitmq.OrderExchange, rabbitmq.StockReservedEvent, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         body,
	}); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	} else {
		log.Printf("%s: reserved %s", op, orderID)
	}

	return reservation, nil
}

func (s *stockService) ChangeStatus(ctx context.Context, reservationID uuid.UUID, status string) error {
	const op = "stock.services.ChangeStatus"

	err := s.tx.WithTx(ctx, func(tx tx.Tx) error {
		if err := s.spotReservationRepo.Update(ctx, tx, reservationID, status); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	event := events.StockStatusChangedEvent{
		ReservationID: reservationID.String(),
		Status:        status,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.channel.PublishWithContext(ctx, rabbitmq.OrderExchange, rabbitmq.StockStatusChangedEvent, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         body,
	}); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	} else {
		log.Printf("%s: changed %s", op, reservationID)
	}

	return nil
}
