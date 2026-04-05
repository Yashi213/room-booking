package mapper

import (
	"github.com/Yashi213/room-booking/internal/api"
	"github.com/Yashi213/room-booking/internal/domain"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func ToCreateScheduleDomain(
	dto api.PostRoomsRoomIdScheduleCreateJSONRequestBody,
	roomID api.RoomIdPath,
) domain.ScheduleCreateInput {
	return domain.ScheduleCreateInput{
		RoomID:     uuid.UUID(roomID),
		DaysOfWeek: dto.DaysOfWeek,
		StartTime:  dto.StartTime,
		EndTime:    dto.EndTime,
	}
}

func ToScheduleDTO(scheduleDomain domain.Schedule) api.Schedule {
	id := openapi_types.UUID(scheduleDomain.ID)
	return api.Schedule{
		Id:         &id,
		RoomId:     openapi_types.UUID(scheduleDomain.RoomID),
		DaysOfWeek: scheduleDomain.DaysOfWeek,
		StartTime:  scheduleDomain.StartTime,
		EndTime:    scheduleDomain.EndTime,
	}
}
