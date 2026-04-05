package handler

import (
	"errors"
	"net/http"

	"github.com/Yashi213/room-booking/internal/service"
)

type mappedAPIError struct {
	status  int
	code    string
	message string
}

func mapServiceError(err error) (mappedAPIError, bool) {
	switch {
	case errors.Is(err, service.ErrForbidden):
		return mappedAPIError{
			status:  http.StatusForbidden,
			code:    "FORBIDDEN",
			message: "admin role required",
		}, true
	case errors.Is(err, service.ErrInvalidScheduleInput):
		return mappedAPIError{
			status:  http.StatusBadRequest,
			code:    "INVALID_REQUEST",
			message: "invalid schedule input",
		}, true
	case errors.Is(err, service.ErrInvalidRoomInput):
		return mappedAPIError{
			status:  http.StatusBadRequest,
			code:    "INVALID_REQUEST",
			message: "invalid room input",
		}, true
	case errors.Is(err, service.ErrInvalidBookingInput):
		return mappedAPIError{
			status:  http.StatusBadRequest,
			code:    "INVALID_REQUEST",
			message: "invalid booking input",
		}, true
	case errors.Is(err, service.ErrUserRoleRequired):
		return mappedAPIError{
			status:  http.StatusForbidden,
			code:    "FORBIDDEN",
			message: "user role required",
		}, true
	case errors.Is(err, service.ErrRoomNotFound):
		return mappedAPIError{
			status:  http.StatusNotFound,
			code:    "ROOM_NOT_FOUND",
			message: "room not found",
		}, true
	case errors.Is(err, service.ErrSlotNotFound):
		return mappedAPIError{
			status:  http.StatusNotFound,
			code:    "SLOT_NOT_FOUND",
			message: "slot not found",
		}, true
	case errors.Is(err, service.ErrBookingNotFound):
		return mappedAPIError{
			status:  http.StatusNotFound,
			code:    "BOOKING_NOT_FOUND",
			message: "booking not found",
		}, true
	case errors.Is(err, service.ErrScheduleExists):
		return mappedAPIError{
			status:  http.StatusConflict,
			code:    "SCHEDULE_EXISTS",
			message: "schedule for this room already exists",
		}, true
	case errors.Is(err, service.ErrSlotAlreadyBooked):
		return mappedAPIError{
			status:  http.StatusConflict,
			code:    "SLOT_ALREADY_BOOKED",
			message: "slot is already booked",
		}, true
	case errors.Is(err, service.ErrInvalidBookingsListParams):
		return mappedAPIError{
			status:  http.StatusBadRequest,
			code:    "INVALID_REQUEST",
			message: "invalid pagination parameters",
		}, true
	case errors.Is(err, service.ErrCannotCancelAnotherUsersBooking):
		return mappedAPIError{
			status:  http.StatusForbidden,
			code:    "FORBIDDEN",
			message: "cannot cancel another user's booking",
		}, true
	case errors.Is(err, service.ErrEmailAlreadyExists):
		return mappedAPIError{
			status:  http.StatusBadRequest,
			code:    "INVALID_REQUEST",
			message: "email already registered",
		}, true
	case errors.Is(err, service.ErrInvalidRegisterInput):
		return mappedAPIError{
			status:  http.StatusBadRequest,
			code:    "INVALID_REQUEST",
			message: "invalid registration data",
		}, true
	case errors.Is(err, service.ErrInvalidCredentials):
		return mappedAPIError{
			status:  http.StatusUnauthorized,
			code:    "UNAUTHORIZED",
			message: "invalid email or password",
		}, true
	case errors.Is(err, service.ErrConferenceServiceUnavailable):
		return mappedAPIError{
			status:  http.StatusServiceUnavailable,
			code:    "CONFERENCE_UNAVAILABLE",
			message: "conference service temporarily unavailable",
		}, true
	default:
		return mappedAPIError{}, false
	}
}
