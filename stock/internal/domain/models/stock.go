package models

import (
	"time"

	"github.com/google/uuid"
)

type ParkingPlaces struct {
	ID             uuid.UUID
	TotalSpots     int
	AvailableSpots int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ParkingSpot struct {
	ID              uuid.UUID
	ParkingLotID    uuid.UUID
	PickUpLocation  string
	DropOffLocation string
}

type Reservation struct {
	ID            uuid.UUID
	OrderID       uuid.UUID
	ExpiresAt     time.Time
	CreatedAt     time.Time
	ParkingSpotID uuid.UUID
	StartsAt      time.Time
	EndsAt        time.Time
	Status        string
}
