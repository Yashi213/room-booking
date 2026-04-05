package mapper

import (
	"github.com/Yashi213/room-booking/internal/api"
	"github.com/Yashi213/room-booking/internal/domain"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func ToRoomDTO(roomDomain domain.Room) api.Room {
	return api.Room{
		Id:          openapi_types.UUID(roomDomain.ID),
		Name:        roomDomain.Name,
		Description: roomDomain.Description,
		Capacity:    roomDomain.Capacity,
		CreatedAt:   roomDomain.CreatedAt,
	}
}

func ToRoomDTOList(roomsDomain []domain.Room) []api.Room {
	result := make([]api.Room, 0, len(roomsDomain))
	for _, roomDomain := range roomsDomain {
		result = append(result, ToRoomDTO(roomDomain))
	}
	return result
}

func ToCreateRoomDomain(dto api.PostRoomsCreateJSONBody) domain.RoomCreateInput {
	return domain.RoomCreateInput{
		Name:        dto.Name,
		Description: dto.Description,
		Capacity:    dto.Capacity,
	}
}
