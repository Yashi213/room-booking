package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Yashi213/room-booking/internal/api"
	"github.com/Yashi213/room-booking/internal/mapper"
)

// GetRoomsRoomIdSlotsList слоты переговорки на дату (admin и user).
// @Summary Слоты по переговорке и дате
// @Tags Slots
// @Produce json
// @Param roomId path string true "UUID переговорки" format(uuid)
// @Param date query string true "Дата YYYY-MM-DD" example(2024-06-10)
// @Success 200 {object} map[string][]api.Slot
// @Failure 400 {object} api.ErrorResponse
// @Failure 401 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Security BearerAuth
// @Router /rooms/{roomId}/slots/list [get]
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
