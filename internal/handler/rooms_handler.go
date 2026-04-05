package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Yashi213/room-booking/internal/api"
	"github.com/Yashi213/room-booking/internal/domain"
	"github.com/Yashi213/room-booking/internal/mapper"
)

func (h *Handler) GetRoomsList(w http.ResponseWriter, r *http.Request) {
	roomsDomain, err := h.roomsService.ListRooms(r.Context())
	if err != nil {
		writeInternalError(w)
		return
	}
	roomDTOs := mapper.ToRoomDTOList(roomsDomain)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"rooms": roomDTOs,
	})
}

func (h *Handler) PostRoomsCreate(w http.ResponseWriter, r *http.Request) {
	role, _, err := h.jwtAccess.ParseBearerAuthorization(r.Header.Get("Authorization"))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	var req api.PostRoomsCreateJSONBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	createRoomDomain := mapper.ToCreateRoomDomain(req)

	roomDomain, err := h.roomsService.CreateRoom(r.Context(), createRoomDomain, domain.UserRole(role))
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
		"room": mapper.ToRoomDTO(roomDomain),
	})
}
