package service

import "errors"

var (
	// Общие
	ErrForbidden = errors.New("forbidden")

	// Комнаты
	ErrRoomNotFound     = errors.New("room not found")
	ErrInvalidRoomInput = errors.New("invalid room input")

	// Расписания
	ErrInvalidScheduleInput = errors.New("invalid schedule input")
	ErrScheduleExists       = errors.New("schedule exists")

	// Бронирования
	ErrInvalidBookingInput             = errors.New("invalid booking input")
	ErrSlotNotFound                    = errors.New("slot not found")
	ErrSlotAlreadyBooked               = errors.New("slot already booked")
	ErrBookingNotFound                 = errors.New("booking not found")
	ErrCannotCancelAnotherUsersBooking = errors.New("cannot cancel another user's booking")
	ErrUserRoleRequired                = errors.New("user role required")
	ErrInvalidBookingsListParams       = errors.New("invalid bookings list params")
	ErrConferenceServiceUnavailable    = errors.New("conference service unavailable")

	// Регистрация / вход
	ErrEmailAlreadyExists   = errors.New("email already exists")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrInvalidRegisterInput = errors.New("invalid register input")
)
