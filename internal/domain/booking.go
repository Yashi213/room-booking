package domain

import (
	"time"

	"github.com/google/uuid"
)

type BookingStatus string

const (
	BookingStatusActive    BookingStatus = "active"
	BookingStatusCancelled BookingStatus = "cancelled"
)

type Booking struct {
	ID             uuid.UUID
	SlotID         uuid.UUID
	UserID         uuid.UUID
	Status         BookingStatus
	ConferenceLink *string
	CreatedAt      *time.Time
}

type BookingCreateInput struct {
	SlotID               uuid.UUID
	UserID               uuid.UUID
	ActorRole            UserRole
	CreateConferenceLink bool
	ConferenceLink       *string
}

type BookingsListInput struct {
	ActorRole UserRole
	Page      int
	PageSize  int
}

type MyBookingsInput struct {
	UserID    uuid.UUID
	ActorRole UserRole
}

type CancelBookingInput struct {
	BookingID uuid.UUID
	UserID    uuid.UUID
	ActorRole UserRole
}
