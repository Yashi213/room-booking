package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Yashi213/room-booking/internal/api"
	"github.com/Yashi213/room-booking/internal/mapper"
)

func (h *Handler) GetRoomsRoomIdSlotsList(
	w http.ResponseWriter,
	r *http.Request,
	roomId api.RoomIdPath,
	params api.GetRoomsRoomIdSlotsListParams,
) {
	if _, _, err := h.jwtAccess.ParseBearerAuthorization(r.Header.Get("Authorization")); err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	listInput := mapper.ToSlotsListInput(roomId, params)
	slotsDomain, err := h.slotsService.ListSlots(r.Context(), listInput)
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
		"slots": mapper.ToSlotDTOList(slotsDomain),
	})
}
