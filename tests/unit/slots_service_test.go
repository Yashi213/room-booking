package unit

import (
	"context"
	"testing"
	"time"

	"github.com/Yashi213/room-booking/internal/domain"
	"github.com/Yashi213/room-booking/internal/mocks"
	"github.com/Yashi213/room-booking/internal/repository"
	"github.com/Yashi213/room-booking/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSlotsService_ListSlots_RoomNotFound(t *testing.T) {
	rid := uuid.New()
	rooms := mocks.NewRoomsRepository(t)
	rooms.EXPECT().GetRoomByID(mock.Anything, rid).Return(domain.Room{}, repository.ErrNotFound)

	svc := service.NewSlotsService(rooms, mocks.NewSlotsRepository(t), mocks.NewSchedulesRepository(t))
	_, err := svc.ListSlots(context.Background(), domain.SlotsListInput{RoomID: rid, Date: time.Now()})
	require.ErrorIs(t, err, service.ErrRoomNotFound)
}

func TestSlotsService_ListSlots_NoSchedule_DelegatesToSlotsRepo(t *testing.T) {
	rid := uuid.New()
	day := time.Now().UTC().Add(24 * time.Hour).Truncate(24 * time.Hour)
	want := []domain.Slot{{ID: uuid.New(), RoomID: rid}}

	rooms := mocks.NewRoomsRepository(t)
	rooms.EXPECT().GetRoomByID(mock.Anything, rid).Return(domain.Room{ID: rid}, nil)

	sched := mocks.NewSchedulesRepository(t)
	sched.EXPECT().GetScheduleByRoomID(mock.Anything, rid).Return(domain.Schedule{}, repository.ErrNotFound)

	slots := mocks.NewSlotsRepository(t)
	slots.EXPECT().ListSlots(mock.Anything, mock.MatchedBy(func(in domain.SlotsListInput) bool {
		return in.RoomID == rid
	})).Return(want, nil)

	svc := service.NewSlotsService(rooms, slots, sched)
	out, err := svc.ListSlots(context.Background(), domain.SlotsListInput{RoomID: rid, Date: day})
	require.NoError(t, err)
	require.Equal(t, want, out)
}

// scheduleDayForTests возвращает дату (UTC полночь) через несколько дней и номер дня недели 1–7 как в CreateSchedule.
func scheduleDayForTests() (day time.Time, dow int) {
	d := time.Now().UTC().Add(10 * 24 * time.Hour).Truncate(24 * time.Hour)
	switch d.Weekday() {
	case time.Sunday:
		return d, 7
	default:
		return d, int(d.Weekday())
	}
}

func TestSlotsService_ListSlots_GeneratesWhenEmpty(t *testing.T) {
	rid := uuid.New()
	day, dow := scheduleDayForTests()

	rooms := mocks.NewRoomsRepository(t)
	rooms.EXPECT().GetRoomByID(mock.Anything, rid).Return(domain.Room{ID: rid}, nil)

	schedRepo := mocks.NewSchedulesRepository(t)
	schedRepo.EXPECT().GetScheduleByRoomID(mock.Anything, rid).Return(domain.Schedule{
		RoomID:     rid,
		DaysOfWeek: []int{dow},
		StartTime:  "09:00",
		EndTime:    "10:00",
	}, nil)

	slotsRepo := mocks.NewSlotsRepository(t)
	slotsRepo.EXPECT().ListSlots(mock.Anything, mock.MatchedBy(func(in domain.SlotsListInput) bool {
		return in.RoomID == rid
	})).Return([]domain.Slot{}, nil).Once()
	slotsRepo.EXPECT().EnsureSlot(mock.Anything, mock.Anything).Return(nil).Times(2)
	slotsRepo.EXPECT().ListSlots(mock.Anything, mock.MatchedBy(func(in domain.SlotsListInput) bool {
		return in.RoomID == rid
	})).Return([]domain.Slot{{ID: uuid.New(), RoomID: rid}}, nil).Once()

	svc := service.NewSlotsService(rooms, slotsRepo, schedRepo)
	out, err := svc.ListSlots(context.Background(), domain.SlotsListInput{RoomID: rid, Date: day})
	require.NoError(t, err)
	require.Len(t, out, 1)
}

func TestSlotsService_CreateSlots_OneHourWindow(t *testing.T) {
	rid := uuid.New()
	mon := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC) // Monday
	sch := domain.Schedule{
		RoomID:     rid,
		DaysOfWeek: []int{1},
		StartTime:  "09:00",
		EndTime:    "10:00",
	}

	slotsRepo := mocks.NewSlotsRepository(t)
	slotsRepo.EXPECT().EnsureSlot(mock.Anything, mock.Anything).Return(nil).Times(2)

	svc := service.NewSlotsService(mocks.NewRoomsRepository(t), slotsRepo, mocks.NewSchedulesRepository(t))
	err := svc.CreateSlots(context.Background(), sch, []time.Time{mon})
	require.NoError(t, err)
}
