package domain

import (
	"time"

	"github.com/google/uuid"
)

type Slot struct {
	ID     uuid.UUID
	RoomID uuid.UUID
	Start  time.Time
	End    time.Time
}

type SlotsListInput struct {
	RoomID uuid.UUID
	Date   time.Time
}

type SlotCreateInput struct {
	RoomID uuid.UUID
	Start  time.Time
	End    time.Time
}
