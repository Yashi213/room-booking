package domain

import "github.com/google/uuid"

type Schedule struct {
	ID         uuid.UUID
	RoomID     uuid.UUID
	DaysOfWeek []int
	StartTime  string
	EndTime    string
}

type ScheduleCreateInput struct {
	RoomID     uuid.UUID
	DaysOfWeek []int
	StartTime  string
	EndTime    string
}
