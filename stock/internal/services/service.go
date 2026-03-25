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

var ErrNoFreeSpots = errors.New("no free spots")

type StockService interface {
	StartReservationCleaner(ctx context.Context)
	Reserve(ctx context.Context, lotID string, orderID string, from time.Time, to time.Time) (*models.Reservation, error)
	ChangeStatus(ctx context.Context, reservationID string, status string) error
	GetAvailability(ctx context.Context, lotID uuid.UUID, from time.Time, to time.Time) (bool, error)
	Release(ctx context.Context, reservationID string) error
	GetReservation(ctx context.Context, orderID string) (string, error)
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

func (s *stockService) StartReservationCleaner(ctx context.Context) {

	ticker := time.NewTicker(1 * time.Minute)

	go func() {
		for {
			select {

			case <-ticker.C:

				reservations, err := s.spotReservationRepo.GetExpired(ctx)
				if err != nil {
					log.Println("reservation cleaner error:", err)
					continue
				}

				for _, r := range reservations {
					err := s.Release(ctx, r.ID)
					if err != nil {
						log.Println("release failed:", err)
					}
				}

			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *stockService) GetAvailability(ctx context.Context, lotID uuid.UUID, from time.Time, to time.Time) (bool, error) {

	available, err := s.parkingSpotRepo.Get(ctx, lotID, from, to)
	if err != nil {
		return false, err
	}

	return available, nil
}

func (s *stockService) Reserve(ctx context.Context, lotID string, orderID string, from time.Time, to time.Time) (*models.Reservation, error) {
	const op = "stock.services.Reserve"

	lotUUID, err := uuid.Parse(lotID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid lotID: %w", op, err)
	}

	spotID, err := s.spotReservationRepo.GetSpot(ctx, lotUUID, from, to)
	if err != nil {
		if errors.Is(err, ErrNoFreeSpots) {
			event := events.StockReservationFailedEvent{
				OrderID: orderID,
				Reason:  err.Error(),
			}

			body, err := json.Marshal(event)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}

			err = s.channel.PublishWithContext(ctx, rabbitmq.OrderExchange, rabbitmq.StockReservationFailedEvent, false, false, amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "application/json",
				Body:         body,
			})

			log.Printf("%s: no free spot for order %s", op, orderID)

			return nil, err
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	spotUUID, err := uuid.Parse(spotID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid spotID: %w", op, err)
	}

	var reservation *models.Reservation
	var reservationID string

	err = s.tx.WithTx(ctx, func(tx tx.Tx) error {

		reservation = &models.Reservation{
			ParkingSpotID: spotUUID,
			OrderID:       orderID,
			CreatedAt:     time.Now(),
			ExpiresAt:     time.Now().Add(15 * time.Minute),
			StartsAt:      from,
			EndsAt:        to,
			Status:        "pending",
		}

		resID, err := s.spotReservationRepo.Create(ctx, tx, reservation)
		if err != nil {
			return err
		}
		reservationID = resID

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// todo: implement price calculation "calculatePrice()"
	event := events.StockReservedEvent{
		ReservationID: reservationID,
		OrderID:       orderID,
		Status:        reservation.Status,
		Amount:        "2",
		Currency:      "RUB",
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

func (s *stockService) ChangeStatus(ctx context.Context, reservationID string, status string) error {
	const op = "stock.services.ChangeStatus"

	err := s.tx.WithTx(ctx, func(tx tx.Tx) error {
		if err := s.spotReservationRepo.Update(ctx, tx, reservationID, status); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	event := events.StockStatusChangedEvent{
		ReservationID: reservationID,
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

func (s *stockService) Release(
	ctx context.Context,
	reservationID string,
) error {

	const op = "stock.services.Release"

	err := s.tx.WithTx(ctx, func(tx tx.Tx) error {
		return s.spotReservationRepo.Update(ctx, tx, reservationID, "released")
	})

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	event := events.StockReleasedEvent{
		ReservationID: reservationID,
		Reason:        "released",
	}

	body, _ := json.Marshal(event)

	err = s.channel.PublishWithContext(
		ctx,
		rabbitmq.OrderExchange,
		rabbitmq.StockReleasedEvent,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)

	if err != nil {
		return fmt.Errorf("%s: publish failed: %w", op, err)
	}

	log.Printf("reservation released %s", reservationID)

	return nil
}

func (s *stockService) GetReservation(ctx context.Context, orderID string) (string, error) {
	orderUUID, err := uuid.Parse(orderID)
	if err != nil {
		return "", err
	}

	return s.spotReservationRepo.GetReservation(ctx, orderUUID)
}
