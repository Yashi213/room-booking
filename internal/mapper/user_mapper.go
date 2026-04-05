package mapper

import (
	"github.com/Yashi213/room-booking/internal/api"
	"github.com/Yashi213/room-booking/internal/domain"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func ToUserDTO(u domain.User) api.User {
	return api.User{
		Id:        openapi_types.UUID(u.ID),
		Email:     openapi_types.Email(u.Email),
		Role:      api.UserRole(u.Role),
		CreatedAt: &u.CreatedAt,
	}
}

func ToUserRegisterInput(dto api.PostRegisterJSONBody) domain.UserRegisterInput {
	return domain.UserRegisterInput{
		Email:    string(dto.Email),
		Password: dto.Password,
		Role:     domain.UserRole(dto.Role),
	}
}

func ToUserLoginInput(dto api.PostLoginJSONBody) domain.UserLoginInput {
	return domain.UserLoginInput{
		Email:    string(dto.Email),
		Password: dto.Password,
	}
}
