package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Yashi213/room-booking/internal/api"
	"github.com/Yashi213/room-booking/internal/domain"
	"github.com/Yashi213/room-booking/internal/mapper"
)

func (h *Handler) PostRoomsRoomIdScheduleCreate(w http.ResponseWriter, r *http.Request, roomId api.RoomIdPath) {
	role, _, err := h.jwtAccess.ParseBearerAuthorization(r.Header.Get("Authorization"))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	var req api.PostRoomsRoomIdScheduleCreateJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	scheduleInput := mapper.ToCreateScheduleDomain(req, roomId)
	scheduleDomain, err := h.schedulesService.CreateSchedule(r.Context(), scheduleInput, domain.UserRole(role))
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
		"schedule": mapper.ToScheduleDTO(scheduleDomain),
	})
}
