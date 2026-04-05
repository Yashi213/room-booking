package mapper

import (
	"github.com/Yashi213/room-booking/internal/api"
	"github.com/Yashi213/room-booking/internal/domain"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func ToSlotsListInput(roomID api.RoomIdPath, params api.GetRoomsRoomIdSlotsListParams) domain.SlotsListInput {
	return domain.SlotsListInput{
		RoomID: uuid.UUID(roomID),
		Date:   params.Date.Time,
	}
}

func ToSlotDTO(slotDomain domain.Slot) api.Slot {
	return api.Slot{
		Id:     openapi_types.UUID(slotDomain.ID),
		RoomId: openapi_types.UUID(slotDomain.RoomID),
		Start:  slotDomain.Start,
		End:    slotDomain.End,
	}
}

func ToSlotDTOList(slotsDomain []domain.Slot) []api.Slot {
	result := make([]api.Slot, 0, len(slotsDomain))
	for _, slotDomain := range slotsDomain {
		result = append(result, ToSlotDTO(slotDomain))
	}
	return result
}
