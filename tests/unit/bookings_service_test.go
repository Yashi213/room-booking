package unit

import (
	"context"
	"errors"
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

func TestBookingsService_CreateBooking_SuccessWithoutConference(t *testing.T) {
	ctx := context.Background()
	slotID := uuid.New()
	userID := uuid.New()
	future := time.Now().UTC().Add(time.Hour)

	slotsRepo := mocks.NewSlotsRepository(t)
	bookingsRepo := mocks.NewBookingsRepository(t)

	slotsRepo.EXPECT().GetSlotByID(mock.Anything, slotID).Return(domain.Slot{ID: slotID, Start: future}, nil)
	bookingsRepo.EXPECT().CreateBooking(mock.Anything, mock.MatchedBy(func(in domain.BookingCreateInput) bool {
		return in.SlotID == slotID && in.UserID == userID && in.ActorRole == domain.RoleUser && !in.CreateConferenceLink && in.ConferenceLink == nil
	})).Return(&domain.Booking{ID: uuid.New(), SlotID: slotID, UserID: userID, Status: domain.BookingStatusActive}, nil)

	svc := service.NewBookingsService(bookingsRepo, slotsRepo, nil, time.Second)
	out, err := svc.CreateBooking(ctx, domain.BookingCreateInput{
		SlotID:    slotID,
		UserID:    userID,
		ActorRole: domain.RoleUser,
	})
	require.NoError(t, err)
	require.Equal(t, slotID, out.SlotID)
}

func TestBookingsService_CreateBooking_WithConference(t *testing.T) {
	ctx := context.Background()
	slotID := uuid.New()
	userID := uuid.New()
	future := time.Now().UTC().Add(time.Hour)
	link := "https://conf.example/join/1"

	slotsRepo := mocks.NewSlotsRepository(t)
	bookingsRepo := mocks.NewBookingsRepository(t)
	conf := mocks.NewMockConferenceService(t)

	slotsRepo.EXPECT().GetSlotByID(mock.Anything, slotID).Return(domain.Slot{ID: slotID, Start: future}, nil)
	conf.EXPECT().CreateConferenceLink(mock.Anything, slotID, userID).Return(link, nil)
	bookingsRepo.EXPECT().CreateBooking(mock.Anything, mock.MatchedBy(func(in domain.BookingCreateInput) bool {
		return in.CreateConferenceLink && in.ConferenceLink != nil && *in.ConferenceLink == link
	})).Return(&domain.Booking{ID: uuid.New(), SlotID: slotID, UserID: userID}, nil)

	svc := service.NewBookingsService(bookingsRepo, slotsRepo, conf, time.Second)
	_, err := svc.CreateBooking(ctx, domain.BookingCreateInput{
		SlotID:               slotID,
		UserID:               userID,
		ActorRole:            domain.RoleUser,
		CreateConferenceLink: true,
	})
	require.NoError(t, err)
}

func TestBookingsService_CreateBooking_Conflict_CallsCancelConference(t *testing.T) {
	ctx := context.Background()
	slotID := uuid.New()
	userID := uuid.New()
	future := time.Now().UTC().Add(time.Hour)
	link := "https://conf.example/x"

	slotsRepo := mocks.NewSlotsRepository(t)
	bookingsRepo := mocks.NewBookingsRepository(t)
	conf := mocks.NewMockConferenceService(t)

	slotsRepo.EXPECT().GetSlotByID(mock.Anything, slotID).Return(domain.Slot{ID: slotID, Start: future}, nil)
	conf.EXPECT().CreateConferenceLink(mock.Anything, slotID, userID).Return(link, nil)
	bookingsRepo.EXPECT().CreateBooking(mock.Anything, mock.Anything).Return((*domain.Booking)(nil), repository.ErrConflict)
	conf.EXPECT().CancelConferenceLink(mock.Anything, link).Return(nil)

	svc := service.NewBookingsService(bookingsRepo, slotsRepo, conf, time.Second)
	_, err := svc.CreateBooking(ctx, domain.BookingCreateInput{
		SlotID:               slotID,
		UserID:               userID,
		ActorRole:            domain.RoleUser,
		CreateConferenceLink: true,
	})
	require.ErrorIs(t, err, service.ErrSlotAlreadyBooked)
}

func TestBookingsService_CreateBooking_NotUserRole(t *testing.T) {
	svc := service.NewBookingsService(nil, nil, nil, time.Second)
	_, err := svc.CreateBooking(context.Background(), domain.BookingCreateInput{
		SlotID:    uuid.New(),
		UserID:    uuid.New(),
		ActorRole: domain.RoleAdmin,
	})
	require.ErrorIs(t, err, service.ErrUserRoleRequired)
}

func TestBookingsService_CreateBooking_InvalidIDs(t *testing.T) {
	svc := service.NewBookingsService(nil, nil, nil, time.Second)
	_, err := svc.CreateBooking(context.Background(), domain.BookingCreateInput{
		SlotID:    uuid.Nil,
		UserID:    uuid.New(),
		ActorRole: domain.RoleUser,
	})
	require.ErrorIs(t, err, service.ErrInvalidBookingInput)
}

func TestBookingsService_CreateBooking_SlotNotFound(t *testing.T) {
	slotsRepo := mocks.NewSlotsRepository(t)
	slotsRepo.EXPECT().GetSlotByID(mock.Anything, mock.Anything).Return(domain.Slot{}, repository.ErrNotFound)

	svc := service.NewBookingsService(mocks.NewBookingsRepository(t), slotsRepo, nil, time.Second)
	_, err := svc.CreateBooking(context.Background(), domain.BookingCreateInput{
		SlotID:    uuid.New(),
		UserID:    uuid.New(),
		ActorRole: domain.RoleUser,
	})
	require.ErrorIs(t, err, service.ErrSlotNotFound)
}

func TestBookingsService_CreateBooking_SlotInPast(t *testing.T) {
	slotID := uuid.New()
	past := time.Now().UTC().Add(-time.Hour)
	slotsRepo := mocks.NewSlotsRepository(t)
	slotsRepo.EXPECT().GetSlotByID(mock.Anything, slotID).Return(domain.Slot{ID: slotID, Start: past}, nil)

	svc := service.NewBookingsService(mocks.NewBookingsRepository(t), slotsRepo, nil, time.Second)
	_, err := svc.CreateBooking(context.Background(), domain.BookingCreateInput{
		SlotID:    slotID,
		UserID:    uuid.New(),
		ActorRole: domain.RoleUser,
	})
	require.ErrorIs(t, err, service.ErrInvalidBookingInput)
}

func TestBookingsService_CreateBooking_ConferenceNil(t *testing.T) {
	slotID := uuid.New()
	userID := uuid.New()
	future := time.Now().UTC().Add(time.Hour)
	slotsRepo := mocks.NewSlotsRepository(t)
	slotsRepo.EXPECT().GetSlotByID(mock.Anything, slotID).Return(domain.Slot{ID: slotID, Start: future}, nil)

	svc := service.NewBookingsService(mocks.NewBookingsRepository(t), slotsRepo, nil, time.Second)
	_, err := svc.CreateBooking(context.Background(), domain.BookingCreateInput{
		SlotID:               slotID,
		UserID:               userID,
		ActorRole:            domain.RoleUser,
		CreateConferenceLink: true,
	})
	require.ErrorIs(t, err, service.ErrConferenceServiceUnavailable)
}

func TestBookingsService_CreateBooking_ConferenceError(t *testing.T) {
	slotID := uuid.New()
	userID := uuid.New()
	future := time.Now().UTC().Add(time.Hour)
	slotsRepo := mocks.NewSlotsRepository(t)
	conf := mocks.NewMockConferenceService(t)
	slotsRepo.EXPECT().GetSlotByID(mock.Anything, slotID).Return(domain.Slot{ID: slotID, Start: future}, nil)
	conf.EXPECT().CreateConferenceLink(mock.Anything, slotID, userID).Return("", errors.New("down"))

	svc := service.NewBookingsService(mocks.NewBookingsRepository(t), slotsRepo, conf, time.Second)
	_, err := svc.CreateBooking(context.Background(), domain.BookingCreateInput{
		SlotID:               slotID,
		UserID:               userID,
		ActorRole:            domain.RoleUser,
		CreateConferenceLink: true,
	})
	require.ErrorIs(t, err, service.ErrConferenceServiceUnavailable)
}

func TestBookingsService_CreateBooking_EmptyConferenceLink(t *testing.T) {
	slotID := uuid.New()
	userID := uuid.New()
	future := time.Now().UTC().Add(time.Hour)
	slotsRepo := mocks.NewSlotsRepository(t)
	conf := mocks.NewMockConferenceService(t)
	slotsRepo.EXPECT().GetSlotByID(mock.Anything, slotID).Return(domain.Slot{ID: slotID, Start: future}, nil)
	conf.EXPECT().CreateConferenceLink(mock.Anything, slotID, userID).Return("   ", nil)

	svc := service.NewBookingsService(mocks.NewBookingsRepository(t), slotsRepo, conf, time.Second)
	_, err := svc.CreateBooking(context.Background(), domain.BookingCreateInput{
		SlotID:               slotID,
		UserID:               userID,
		ActorRole:            domain.RoleUser,
		CreateConferenceLink: true,
	})
	require.ErrorIs(t, err, service.ErrConferenceServiceUnavailable)
}

func TestBookingsService_ListBookings_ForbiddenNonAdmin(t *testing.T) {
	svc := service.NewBookingsService(nil, nil, nil, time.Second)
	_, _, err := svc.ListBookings(context.Background(), domain.BookingsListInput{
		ActorRole: domain.RoleUser,
		Page:      1,
		PageSize:  10,
	})
	require.ErrorIs(t, err, service.ErrForbidden)
}

func TestBookingsService_ListBookings_InvalidPagination(t *testing.T) {
	svc := service.NewBookingsService(nil, nil, nil, time.Second)
	_, _, err := svc.ListBookings(context.Background(), domain.BookingsListInput{
		ActorRole: domain.RoleAdmin,
		Page:      0,
		PageSize:  10,
	})
	require.ErrorIs(t, err, service.ErrInvalidBookingsListParams)
}

func TestBookingsService_ListBookings_Success(t *testing.T) {
	bookingsRepo := mocks.NewBookingsRepository(t)
	bookingsRepo.EXPECT().ListBookings(mock.Anything, mock.MatchedBy(func(in domain.BookingsListInput) bool {
		return in.Page == 1 && in.PageSize == 20
	})).Return([]domain.Booking{{ID: uuid.New()}}, 1, nil)

	svc := service.NewBookingsService(bookingsRepo, nil, nil, time.Second)
	list, total, err := svc.ListBookings(context.Background(), domain.BookingsListInput{
		ActorRole: domain.RoleAdmin,
		Page:      1,
		PageSize:  20,
	})
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, 1, total)
}

func TestBookingsService_ListMyBookings_NotUser(t *testing.T) {
	svc := service.NewBookingsService(nil, nil, nil, time.Second)
	_, err := svc.ListMyBookings(context.Background(), domain.MyBookingsInput{
		UserID:    uuid.New(),
		ActorRole: domain.RoleAdmin,
	})
	require.ErrorIs(t, err, service.ErrUserRoleRequired)
}

func TestBookingsService_CancelBooking_Idempotent(t *testing.T) {
	uid := uuid.New()
	bid := uuid.New()
	bookingsRepo := mocks.NewBookingsRepository(t)
	bookingsRepo.EXPECT().GetBookingByID(mock.Anything, bid).Return(&domain.Booking{
		ID:     bid,
		UserID: uid,
		Status: domain.BookingStatusCancelled,
	}, nil)

	svc := service.NewBookingsService(bookingsRepo, nil, nil, time.Second)
	out, err := svc.CancelBooking(context.Background(), domain.CancelBookingInput{
		BookingID: bid,
		UserID:    uid,
		ActorRole: domain.RoleUser,
	})
	require.NoError(t, err)
	require.Equal(t, domain.BookingStatusCancelled, out.Status)
}

func TestBookingsService_CancelBooking_WrongUser(t *testing.T) {
	bid := uuid.New()
	bookingsRepo := mocks.NewBookingsRepository(t)
	bookingsRepo.EXPECT().GetBookingByID(mock.Anything, bid).Return(&domain.Booking{
		ID:     bid,
		UserID: uuid.New(),
		Status: domain.BookingStatusActive,
	}, nil)

	svc := service.NewBookingsService(bookingsRepo, nil, nil, time.Second)
	_, err := svc.CancelBooking(context.Background(), domain.CancelBookingInput{
		BookingID: bid,
		UserID:    uuid.New(),
		ActorRole: domain.RoleUser,
	})
	require.ErrorIs(t, err, service.ErrCannotCancelAnotherUsersBooking)
}

func TestBookingsService_CancelBooking_Success(t *testing.T) {
	uid := uuid.New()
	bid := uuid.New()
	bookingsRepo := mocks.NewBookingsRepository(t)
	bookingsRepo.EXPECT().GetBookingByID(mock.Anything, bid).Return(&domain.Booking{
		ID:     bid,
		UserID: uid,
		Status: domain.BookingStatusActive,
	}, nil)
	bookingsRepo.EXPECT().UpdateBookingStatus(mock.Anything, bid, domain.BookingStatusCancelled).
		Return(&domain.Booking{ID: bid, UserID: uid, Status: domain.BookingStatusCancelled}, nil)

	svc := service.NewBookingsService(bookingsRepo, nil, nil, time.Second)
	out, err := svc.CancelBooking(context.Background(), domain.CancelBookingInput{
		BookingID: bid,
		UserID:    uid,
		ActorRole: domain.RoleUser,
	})
	require.NoError(t, err)
	require.Equal(t, domain.BookingStatusCancelled, out.Status)
}
