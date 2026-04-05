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

// slotsSvcStub — SlotsService без mockery (интерфейс в пакете service).
type slotsSvcStub struct {
	createErr error
	calls     int
}

func (s *slotsSvcStub) ListSlots(ctx context.Context, _ domain.SlotsListInput) ([]domain.Slot, error) {
	return nil, nil
}

func (s *slotsSvcStub) CreateSlots(ctx context.Context, _ domain.Schedule, _ []time.Time) error {
	s.calls++
	return s.createErr
}

func (s *slotsSvcStub) RefillRollingWindowSlots(ctx context.Context) error {
	return nil
}

func TestSchedulesService_CreateSchedule_Forbidden(t *testing.T) {
	svc := service.NewSchedulesService(nil, nil, &slotsSvcStub{})
	_, err := svc.CreateSchedule(context.Background(), domain.ScheduleCreateInput{
		RoomID:     uuid.New(),
		DaysOfWeek: []int{1},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}, domain.RoleUser)
	require.ErrorIs(t, err, service.ErrForbidden)
}

func TestSchedulesService_CreateSchedule_InvalidDays(t *testing.T) {
	svc := service.NewSchedulesService(nil, nil, &slotsSvcStub{})
	_, err := svc.CreateSchedule(context.Background(), domain.ScheduleCreateInput{
		RoomID:     uuid.New(),
		DaysOfWeek: []int{},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}, domain.RoleAdmin)
	require.ErrorIs(t, err, service.ErrInvalidScheduleInput)
}

func TestSchedulesService_CreateSchedule_InvalidDayRange(t *testing.T) {
	svc := service.NewSchedulesService(nil, nil, &slotsSvcStub{})
	_, err := svc.CreateSchedule(context.Background(), domain.ScheduleCreateInput{
		RoomID:     uuid.New(),
		DaysOfWeek: []int{8},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}, domain.RoleAdmin)
	require.ErrorIs(t, err, service.ErrInvalidScheduleInput)
}

func TestSchedulesService_CreateSchedule_TimeOrder(t *testing.T) {
	svc := service.NewSchedulesService(nil, nil, &slotsSvcStub{})
	_, err := svc.CreateSchedule(context.Background(), domain.ScheduleCreateInput{
		RoomID:     uuid.New(),
		DaysOfWeek: []int{1},
		StartTime:  "18:00",
		EndTime:    "09:00",
	}, domain.RoleAdmin)
	require.ErrorIs(t, err, service.ErrInvalidScheduleInput)
}

func TestSchedulesService_CreateSchedule_RoomNotFound(t *testing.T) {
	rid := uuid.New()
	rooms := mocks.NewRoomsRepository(t)
	rooms.EXPECT().GetRoomByID(mock.Anything, rid).Return(domain.Room{}, repository.ErrNotFound)

	svc := service.NewSchedulesService(rooms, mocks.NewSchedulesRepository(t), &slotsSvcStub{})
	_, err := svc.CreateSchedule(context.Background(), domain.ScheduleCreateInput{
		RoomID:     rid,
		DaysOfWeek: []int{1},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}, domain.RoleAdmin)
	require.ErrorIs(t, err, service.ErrRoomNotFound)
}

func TestSchedulesService_CreateSchedule_AlreadyExists(t *testing.T) {
	rid := uuid.New()
	rooms := mocks.NewRoomsRepository(t)
	rooms.EXPECT().GetRoomByID(mock.Anything, rid).Return(domain.Room{ID: rid}, nil)

	schedRepo := mocks.NewSchedulesRepository(t)
	schedRepo.EXPECT().CreateSchedule(mock.Anything, mock.Anything).Return(domain.Schedule{}, repository.ErrConflict)

	stub := &slotsSvcStub{}
	svc := service.NewSchedulesService(rooms, schedRepo, stub)
	_, err := svc.CreateSchedule(context.Background(), domain.ScheduleCreateInput{
		RoomID:     rid,
		DaysOfWeek: []int{1},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}, domain.RoleAdmin)
	require.ErrorIs(t, err, service.ErrScheduleExists)
	require.Equal(t, 0, stub.calls)
}

func TestSchedulesService_CreateSchedule_CallsCreateSlots(t *testing.T) {
	rid := uuid.New()
	sch := domain.Schedule{ID: uuid.New(), RoomID: rid, DaysOfWeek: []int{1}, StartTime: "09:00", EndTime: "18:00"}

	rooms := mocks.NewRoomsRepository(t)
	rooms.EXPECT().GetRoomByID(mock.Anything, rid).Return(domain.Room{ID: rid}, nil)

	schedRepo := mocks.NewSchedulesRepository(t)
	schedRepo.EXPECT().CreateSchedule(mock.Anything, mock.Anything).Return(sch, nil)

	stub := &slotsSvcStub{}
	svc := service.NewSchedulesService(rooms, schedRepo, stub)
	out, err := svc.CreateSchedule(context.Background(), domain.ScheduleCreateInput{
		RoomID:     rid,
		DaysOfWeek: []int{1},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}, domain.RoleAdmin)
	require.NoError(t, err)
	require.Equal(t, sch.ID, out.ID)
	require.Equal(t, 1, stub.calls) // один вызов CreateSlots со срезом из RollingWindowDays дней
}

func TestSchedulesService_CreateSchedule_CreateSlotsError(t *testing.T) {
	rid := uuid.New()
	sch := domain.Schedule{ID: uuid.New(), RoomID: rid}

	rooms := mocks.NewRoomsRepository(t)
	rooms.EXPECT().GetRoomByID(mock.Anything, rid).Return(domain.Room{ID: rid}, nil)

	schedRepo := mocks.NewSchedulesRepository(t)
	schedRepo.EXPECT().CreateSchedule(mock.Anything, mock.Anything).Return(sch, nil)

	stub := &slotsSvcStub{createErr: context.Canceled}
	svc := service.NewSchedulesService(rooms, schedRepo, stub)
	_, err := svc.CreateSchedule(context.Background(), domain.ScheduleCreateInput{
		RoomID:     rid,
		DaysOfWeek: []int{1},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}, domain.RoleAdmin)
	require.ErrorIs(t, err, context.Canceled)
}
