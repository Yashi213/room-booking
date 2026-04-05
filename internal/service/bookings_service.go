package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Yashi213/room-booking/internal/conference"
	"github.com/Yashi213/room-booking/internal/domain"
	"github.com/Yashi213/room-booking/internal/repository"
	"github.com/google/uuid"
)

type BookingsService interface {
	CreateBooking(ctx context.Context, input domain.BookingCreateInput) (*domain.Booking, error)
	ListBookings(ctx context.Context, input domain.BookingsListInput) ([]domain.Booking, int, error)
	ListMyBookings(ctx context.Context, input domain.MyBookingsInput) ([]domain.Booking, error)
	CancelBooking(ctx context.Context, input domain.CancelBookingInput) (*domain.Booking, error)
}

type bookingsService struct {
	bookingsRepository repository.BookingsRepository
	slotsRepository    repository.SlotsRepository
	conference         conference.Service
	conferenceTimeout  time.Duration
}

func NewBookingsService(
	bookingsRepository repository.BookingsRepository,
	slotsRepository repository.SlotsRepository,
	conf conference.Service,
	conferenceTimeout time.Duration,
) BookingsService {
	if conferenceTimeout <= 0 {
		conferenceTimeout = 3 * time.Second
	}
	return &bookingsService{
		bookingsRepository: bookingsRepository,
		slotsRepository:    slotsRepository,
		conference:         conf,
		conferenceTimeout:  conferenceTimeout,
	}
}

func (s *bookingsService) CreateBooking(ctx context.Context, input domain.BookingCreateInput) (*domain.Booking, error) {
	if input.ActorRole != domain.RoleUser {
		return nil, ErrUserRoleRequired
	}
	if input.SlotID == uuid.Nil || input.UserID == uuid.Nil {
		return nil, ErrInvalidBookingInput
	}

	slot, err := s.slotsRepository.GetSlotByID(ctx, input.SlotID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSlotNotFound
		}
		return nil, err
	}
	if slot.Start.Before(time.Now().UTC()) {
		return nil, ErrInvalidBookingInput
	}

	if input.CreateConferenceLink {
		if s.conference == nil {
			return nil, ErrConferenceServiceUnavailable
		}
		cctx, cancel := context.WithTimeout(ctx, s.conferenceTimeout)
		link, err := s.conference.CreateConferenceLink(cctx, input.SlotID, input.UserID)
		cancel()
		if err != nil {
			return nil, ErrConferenceServiceUnavailable
		}
		link = strings.TrimSpace(link)
		if link == "" {
			return nil, ErrConferenceServiceUnavailable
		}
		input.ConferenceLink = &link
	}

	booking, err := s.bookingsRepository.CreateBooking(ctx, input)
	if err != nil {
		if input.CreateConferenceLink && input.ConferenceLink != nil && s.conference != nil {
			if link := strings.TrimSpace(*input.ConferenceLink); link != "" {
				rctx, rcancel := context.WithTimeout(context.Background(), s.conferenceTimeout)
				_ = s.conference.CancelConferenceLink(rctx, link)
				rcancel()
			}
		}
		if errors.Is(err, repository.ErrConflict) {
			return nil, ErrSlotAlreadyBooked
		}
		return nil, err
	}

	return booking, nil
}

func (s *bookingsService) ListBookings(ctx context.Context, input domain.BookingsListInput) ([]domain.Booking, int, error) {
	if input.ActorRole != domain.RoleAdmin {
		return nil, 0, ErrForbidden
	}
	if input.Page < 1 || input.PageSize < 1 || input.PageSize > 100 {
		return nil, 0, ErrInvalidBookingsListParams
	}
	return s.bookingsRepository.ListBookings(ctx, input)
}

func (s *bookingsService) ListMyBookings(ctx context.Context, input domain.MyBookingsInput) ([]domain.Booking, error) {
	if input.ActorRole != domain.RoleUser {
		return nil, ErrUserRoleRequired
	}
	if input.UserID == uuid.Nil {
		return nil, ErrInvalidBookingInput
	}
	return s.bookingsRepository.ListMyBookings(ctx, input.UserID, time.Now().UTC())
}

func (s *bookingsService) CancelBooking(ctx context.Context, input domain.CancelBookingInput) (*domain.Booking, error) {
	if input.ActorRole != domain.RoleUser {
		return nil, ErrUserRoleRequired
	}
	if input.BookingID == uuid.Nil || input.UserID == uuid.Nil {
		return nil, ErrInvalidBookingInput
	}

	booking, err := s.bookingsRepository.GetBookingByID(ctx, input.BookingID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}

	if booking.UserID != input.UserID {
		return nil, ErrCannotCancelAnotherUsersBooking
	}
	if booking.Status == domain.BookingStatusCancelled {
		return booking, nil
	}

	updatedBooking, err := s.bookingsRepository.UpdateBookingStatus(ctx, input.BookingID, domain.BookingStatusCancelled)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}

	return updatedBooking, nil
}
