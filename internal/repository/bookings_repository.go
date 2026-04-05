package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Yashi213/room-booking/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookingsRepository interface {
	CreateBooking(ctx context.Context, input domain.BookingCreateInput) (*domain.Booking, error)
	ListBookings(ctx context.Context, input domain.BookingsListInput) ([]domain.Booking, int, error)
	ListMyBookings(ctx context.Context, userID uuid.UUID, now time.Time) ([]domain.Booking, error)
	GetBookingByID(ctx context.Context, bookingID uuid.UUID) (*domain.Booking, error)
	UpdateBookingStatus(ctx context.Context, bookingID uuid.UUID, status domain.BookingStatus) (*domain.Booking, error)
}

type bookingsRepository struct {
	db *pgxpool.Pool
}

func NewBookingsRepository(db *pgxpool.Pool) BookingsRepository {
	return &bookingsRepository{db: db}
}

func (r *bookingsRepository) CreateBooking(ctx context.Context, input domain.BookingCreateInput) (*domain.Booking, error) {
	var booking domain.Booking

	err := r.db.QueryRow(ctx, `
		INSERT INTO bookings (slot_id, user_id, status_id, conference_link)
		VALUES (
			$1, $2,
			(SELECT s.id FROM booking_statuses s WHERE s.code = $3),
			$4
		)
		RETURNING id, slot_id, user_id,
			(SELECT s.code FROM booking_statuses s WHERE s.id = status_id),
			conference_link, created_at
	`, input.SlotID, input.UserID, string(domain.BookingStatusActive), input.ConferenceLink).
		Scan(&booking.ID, &booking.SlotID, &booking.UserID, &booking.Status, &booking.ConferenceLink, &booking.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrConflict
		}
		return nil, err
	}

	return &booking, nil
}

func (r *bookingsRepository) ListBookings(ctx context.Context, input domain.BookingsListInput) ([]domain.Booking, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM bookings`).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (input.Page - 1) * input.PageSize
	rows, err := r.db.Query(ctx, `
		SELECT b.id, b.slot_id, b.user_id, s.code, b.conference_link, b.created_at
		FROM bookings b
		INNER JOIN booking_statuses s ON s.id = b.status_id
		ORDER BY b.created_at DESC
		LIMIT $1 OFFSET $2
	`, input.PageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	bookings := make([]domain.Booking, 0)
	for rows.Next() {
		var b domain.Booking
		if err := rows.Scan(&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt); err != nil {
			return nil, 0, err
		}
		bookings = append(bookings, b)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return bookings, total, nil
}

func (r *bookingsRepository) ListMyBookings(ctx context.Context, userID uuid.UUID, now time.Time) ([]domain.Booking, error) {
	rows, err := r.db.Query(ctx, `
		SELECT b.id, b.slot_id, b.user_id, st.code, b.conference_link, b.created_at
		FROM bookings b
		INNER JOIN slots sl ON sl.id = b.slot_id
		INNER JOIN booking_statuses st ON st.id = b.status_id
		WHERE b.user_id = $1
		  AND sl.start_time >= $2
		ORDER BY sl.start_time ASC
	`, userID, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookings := make([]domain.Booking, 0)
	for rows.Next() {
		var b domain.Booking
		if err := rows.Scan(&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt); err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bookings, nil
}

func (r *bookingsRepository) GetBookingByID(ctx context.Context, bookingID uuid.UUID) (*domain.Booking, error) {
	var booking domain.Booking

	err := r.db.QueryRow(ctx, `
		SELECT b.id, b.slot_id, b.user_id, s.code, b.conference_link, b.created_at
		FROM bookings b
		INNER JOIN booking_statuses s ON s.id = b.status_id
		WHERE b.id = $1
	`, bookingID).Scan(
		&booking.ID,
		&booking.SlotID,
		&booking.UserID,
		&booking.Status,
		&booking.ConferenceLink,
		&booking.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &booking, nil
}

func (r *bookingsRepository) UpdateBookingStatus(ctx context.Context, bookingID uuid.UUID, status domain.BookingStatus) (*domain.Booking, error) {
	var booking domain.Booking

	err := r.db.QueryRow(ctx, `
		UPDATE bookings
		SET status_id = (SELECT s.id FROM booking_statuses s WHERE s.code = $2)
		WHERE id = $1
		RETURNING id, slot_id, user_id,
			(SELECT s.code FROM booking_statuses s WHERE s.id = status_id),
			conference_link, created_at
	`, bookingID, string(status)).Scan(
		&booking.ID,
		&booking.SlotID,
		&booking.UserID,
		&booking.Status,
		&booking.ConferenceLink,
		&booking.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &booking, nil
}
