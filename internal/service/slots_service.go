package service

import (
	"context"
	"errors"
	"time"

	"github.com/Yashi213/room-booking/internal/domain"
	"github.com/Yashi213/room-booking/internal/repository"
)

const RollingWindowDays = 7

type SlotsService interface {
	ListSlots(ctx context.Context, input domain.SlotsListInput) ([]domain.Slot, error)
	CreateSlots(ctx context.Context, schedule domain.Schedule, days []time.Time) error
	RefillRollingWindowSlots(ctx context.Context) error
}

type slotsService struct {
	roomsRepository     repository.RoomsRepository
	slotsRepository     repository.SlotsRepository
	schedulesRepository repository.SchedulesRepository
}

func NewSlotsService(
	roomsRepository repository.RoomsRepository,
	slotsRepository repository.SlotsRepository,
	schedulesRepository repository.SchedulesRepository,
) SlotsService {
	return &slotsService{
		roomsRepository:     roomsRepository,
		slotsRepository:     slotsRepository,
		schedulesRepository: schedulesRepository,
	}
}

func (s *slotsService) ListSlots(ctx context.Context, input domain.SlotsListInput) ([]domain.Slot, error) {
	if _, err := s.roomsRepository.GetRoomByID(ctx, input.RoomID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRoomNotFound
		}
		return nil, err
	}

	schedule, err := s.schedulesRepository.GetScheduleByRoomID(ctx, input.RoomID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return s.slotsRepository.ListSlots(ctx, input)
		}
		return nil, err
	}

	requestedDay := input.Date.UTC().Truncate(24 * time.Hour)
	today := time.Now().UTC().Truncate(24 * time.Hour)

	if requestedDay.Before(today) {
		return s.slotsRepository.ListSlots(ctx, input)
	}
	if !weekdayInSchedule(schedule.DaysOfWeek, requestedDay.Weekday()) {
		return s.slotsRepository.ListSlots(ctx, input)
	}

	slots, err := s.slotsRepository.ListSlots(ctx, input)
	if err != nil {
		return nil, err
	}
	if len(slots) == 0 {
		if err := s.CreateSlots(ctx, schedule, []time.Time{requestedDay}); err != nil {
			return nil, err
		}
	}

	return s.slotsRepository.ListSlots(ctx, input)
}

func (s *slotsService) CreateSlots(ctx context.Context, schedule domain.Schedule, days []time.Time) error {
	startClock, err := time.Parse("15:04", schedule.StartTime)
	if err != nil {
		return err
	}
	endClock, err := time.Parse("15:04", schedule.EndTime)
	if err != nil {
		return err
	}

	for _, day := range days {
		d := day.UTC().Truncate(24 * time.Hour)
		if !weekdayInSchedule(schedule.DaysOfWeek, d.Weekday()) {
			continue
		}
		windowStart := time.Date(d.Year(), d.Month(), d.Day(), startClock.Hour(), startClock.Minute(), 0, 0, time.UTC)
		windowEnd := time.Date(d.Year(), d.Month(), d.Day(), endClock.Hour(), endClock.Minute(), 0, 0, time.UTC)

		for t := windowStart; t.Before(windowEnd); t = t.Add(30 * time.Minute) {
			end := t.Add(30 * time.Minute)
			if end.After(windowEnd) {
				break
			}
			if err := s.slotsRepository.EnsureSlot(ctx, domain.SlotCreateInput{
				RoomID: schedule.RoomID,
				Start:  t,
				End:    end,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *slotsService) RefillRollingWindowSlots(ctx context.Context) error {
	schedules, err := s.schedulesRepository.ListAllSchedules(ctx)
	if err != nil {
		return err
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	days := make([]time.Time, RollingWindowDays)
	for i := range days {
		days[i] = today.AddDate(0, 0, i)
	}
	for _, sch := range schedules {
		if err := s.CreateSlots(ctx, sch, days); err != nil {
			return err
		}
	}
	return nil
}

func weekdayInSchedule(days []int, w time.Weekday) bool {
	for _, day := range days {
		if time.Weekday(day%7) == w {
			return true
		}
	}
	return false
}
