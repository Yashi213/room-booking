package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Yashi213/room-booking/internal/api"
	"github.com/Yashi213/room-booking/internal/domain"
	"github.com/Yashi213/room-booking/internal/mapper"
)

// GetBookingsList все брони с пагинацией (только admin).
// @Summary Список всех броней
// @Tags Bookings
// @Produce json
// @Param page query int false "Номер страницы" default(1)
// @Param pageSize query int false "Размер страницы" default(20)
// @Success 200 {object} api.GetBookingsList200JSONResponse
// @Failure 401 {object} api.ErrorResponse
// @Failure 403 {object} api.ErrorResponse
// @Security BearerAuth
// @Router /bookings/list [get]
func (h *Handler) GetBookingsList(w http.ResponseWriter, r *http.Request, params api.GetBookingsListParams) {
	role, _, err := h.jwtAccess.ParseBearerAuthorization(r.Header.Get("Authorization"))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	listInput := mapper.ToBookingsListInput(params, domain.UserRole(role))
	bookingsDomain, total, err := h.bookingsService.ListBookings(r.Context(), listInput)
	if err != nil {
		if mappedErr, ok := mapServiceError(err); ok {
			writeError(w, mappedErr.status, mappedErr.code, mappedErr.message)
			return
		}
		writeInternalError(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"bookings":   mapper.ToBookingDTOList(bookingsDomain),
		"pagination": mapper.ToPaginationDTO(listInput.Page, listInput.PageSize, total),
	})
}

// GetBookingsMy брони текущего пользователя на будущие слоты (только user).
// @Summary Мои брони
// @Tags Bookings
// @Produce json
// @Success 200 {object} map[string][]api.Booking
// @Failure 401 {object} api.ErrorResponse
// @Failure 403 {object} api.ErrorResponse
// @Security BearerAuth
// @Router /bookings/my [get]
func (h *Handler) GetBookingsMy(w http.ResponseWriter, r *http.Request) {
	role, userID, err := h.jwtAccess.ParseBearerAuthorization(r.Header.Get("Authorization"))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	listInput := mapper.ToMyBookingsInput(userID, domain.UserRole(role))
	bookingsDomain, err := h.bookingsService.ListMyBookings(r.Context(), listInput)
	if err != nil {
		if mappedErr, ok := mapServiceError(err); ok {
			writeError(w, mappedErr.status, mappedErr.code, mappedErr.message)
			return
		}
		writeInternalError(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"bookings": mapper.ToBookingDTOList(bookingsDomain),
	})
}

// PostBookingsBookingIdCancel отмена своей брони (только user, идемпотентно).
// @Summary Отменить бронь
// @Tags Bookings
// @Produce json
// @Param bookingId path string true "UUID брони" format(uuid)
// @Success 200 {object} map[string]api.Booking
// @Failure 400 {object} api.ErrorResponse
// @Failure 401 {object} api.ErrorResponse
// @Failure 403 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Security BearerAuth
// @Router /bookings/{bookingId}/cancel [post]
func (h *Handler) PostBookingsBookingIdCancel(w http.ResponseWriter, r *http.Request, bookingId api.BookingIdPath) {
	role, userID, err := h.jwtAccess.ParseBearerAuthorization(r.Header.Get("Authorization"))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	cancelInput := mapper.ToCancelBookingInput(bookingId, userID, domain.UserRole(role))
	bookingDomain, err := h.bookingsService.CancelBooking(r.Context(), cancelInput)
	if err != nil {
		if mappedErr, ok := mapServiceError(err); ok {
			writeError(w, mappedErr.status, mappedErr.code, mappedErr.message)
			return
		}
		writeInternalError(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"booking": mapper.ToBookingDTO(*bookingDomain),
	})
}

// PostBookingsCreate бронь на слот (только user).
// @Summary Создать бронь
// @Tags Bookings
// @Accept json
// @Produce json
// @Param body body api.PostBookingsCreateJSONBody true "slotId, createConferenceLink"
// @Success 201 {object} map[string]api.Booking
// @Failure 400 {object} api.ErrorResponse
// @Failure 401 {object} api.ErrorResponse
// @Failure 403 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 409 {object} api.ErrorResponse
// @Security BearerAuth
// @Router /bookings/create [post]
func (h *Handler) PostBookingsCreate(w http.ResponseWriter, r *http.Request) {
	role, userID, err := h.jwtAccess.ParseBearerAuthorization(r.Header.Get("Authorization"))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	var req api.PostBookingsCreateJSONBody

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	createBookingInput := mapper.ToCreateBookingInput(req, userID, domain.UserRole(role))
	bookingDomain, err := h.bookingsService.CreateBooking(r.Context(), createBookingInput)
	if err != nil {
		if mappedErr, ok := mapServiceError(err); ok {
			writeError(w, mappedErr.status, mappedErr.code, mappedErr.message)
			return
		}
		writeInternalError(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"booking": mapper.ToBookingDTO(*bookingDomain),
	})
}
