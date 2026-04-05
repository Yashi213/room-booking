package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Yashi213/room-booking/internal/domain"
	"github.com/Yashi213/room-booking/internal/repository"
)

type SchedulesService interface {
	CreateSchedule(ctx context.Context, input domain.ScheduleCreateInput, actorRole domain.UserRole) (domain.Schedule, error)
}

type schedulesService struct {
	roomsRepo     repository.RoomsRepository
	schedulesRepo repository.SchedulesRepository
	slotsService  SlotsService
}

func NewSchedulesService(
	roomsRepo repository.RoomsRepository,
	schedulesRepo repository.SchedulesRepository,
	slotsService SlotsService,
) SchedulesService {
	return &schedulesService{
		roomsRepo:     roomsRepo,
		schedulesRepo: schedulesRepo,
		slotsService:  slotsService,
	}
}

func (s *schedulesService) CreateSchedule(ctx context.Context, input domain.ScheduleCreateInput, actorRole domain.UserRole) (domain.Schedule, error) {
	if actorRole != domain.RoleAdmin {
		return domain.Schedule{}, ErrForbidden
	}

	if len(input.DaysOfWeek) == 0 {
		return domain.Schedule{}, ErrInvalidScheduleInput
	}
	for _, day := range input.DaysOfWeek {
		if day < 1 || day > 7 {
			return domain.Schedule{}, ErrInvalidScheduleInput
		}
	}
	if strings.TrimSpace(input.StartTime) == "" || strings.TrimSpace(input.EndTime) == "" {
		return domain.Schedule{}, ErrInvalidScheduleInput
	}
	start, err := time.Parse("15:04", input.StartTime)
	if err != nil {
		return domain.Schedule{}, ErrInvalidScheduleInput
	}
	end, err := time.Parse("15:04", input.EndTime)
	if err != nil {
		return domain.Schedule{}, ErrInvalidScheduleInput
	}
	if !start.Before(end) {
		return domain.Schedule{}, ErrInvalidScheduleInput
	}

	if _, err := s.roomsRepo.GetRoomByID(ctx, input.RoomID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.Schedule{}, ErrRoomNotFound
		}
		return domain.Schedule{}, err
	}

	schedule, err := s.schedulesRepo.CreateSchedule(ctx, input)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return domain.Schedule{}, ErrScheduleExists
		}
		return domain.Schedule{}, err
	}

	weekStart := time.Now().UTC().Truncate(24 * time.Hour)
	days := make([]time.Time, RollingWindowDays)
	for i := range days {
		days[i] = weekStart.AddDate(0, 0, i)
	}
	if err = s.slotsService.CreateSlots(ctx, schedule, days); err != nil {
		return domain.Schedule{}, err
	}

	return schedule, nil
}
