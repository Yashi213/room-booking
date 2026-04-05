package mapper

import (
	"github.com/Yashi213/room-booking/internal/api"
	"github.com/Yashi213/room-booking/internal/domain"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func ToCreateBookingInput(dto api.PostBookingsCreateJSONBody, userID uuid.UUID, actorRole domain.UserRole) domain.BookingCreateInput {
	createConferenceLink := false
	if dto.CreateConferenceLink != nil {
		createConferenceLink = *dto.CreateConferenceLink
	}

	return domain.BookingCreateInput{
		SlotID:               uuid.UUID(dto.SlotId),
		UserID:               userID,
		ActorRole:            actorRole,
		CreateConferenceLink: createConferenceLink,
	}
}

func ToBookingsListInput(params api.GetBookingsListParams, actorRole domain.UserRole) domain.BookingsListInput {
	page := 1
	if params.Page != nil {
		page = *params.Page
	}
	pageSize := 20
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}
	return domain.BookingsListInput{
		ActorRole: actorRole,
		Page:      page,
		PageSize:  pageSize,
	}
}

func ToMyBookingsInput(userID uuid.UUID, actorRole domain.UserRole) domain.MyBookingsInput {
	return domain.MyBookingsInput{
		UserID:    userID,
		ActorRole: actorRole,
	}
}

func ToCancelBookingInput(bookingID api.BookingIdPath, userID uuid.UUID, actorRole domain.UserRole) domain.CancelBookingInput {
	return domain.CancelBookingInput{
		BookingID: uuid.UUID(bookingID),
		UserID:    userID,
		ActorRole: actorRole,
	}
}

func ToBookingDTOList(bookingsDomain []domain.Booking) []api.Booking {
	result := make([]api.Booking, 0, len(bookingsDomain))
	for _, b := range bookingsDomain {
		result = append(result, ToBookingDTO(b))
	}
	return result
}

func ToPaginationDTO(page, pageSize, total int) api.Pagination {
	return api.Pagination{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}
}

func ToBookingDTO(bookingDomain domain.Booking) api.Booking {
	return api.Booking{
		Id:             openapi_types.UUID(bookingDomain.ID),
		SlotId:         openapi_types.UUID(bookingDomain.SlotID),
		UserId:         openapi_types.UUID(bookingDomain.UserID),
		Status:         api.BookingStatus(bookingDomain.Status),
		ConferenceLink: bookingDomain.ConferenceLink,
		CreatedAt:      bookingDomain.CreatedAt,
	}
}
