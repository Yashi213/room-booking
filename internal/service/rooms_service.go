package service

import (
	"context"
	"strings"

	"github.com/Yashi213/room-booking/internal/domain"
	"github.com/Yashi213/room-booking/internal/repository"
)

type RoomsService interface {
	ListRooms(ctx context.Context) ([]domain.Room, error)
	CreateRoom(ctx context.Context, input domain.RoomCreateInput, actorRole domain.UserRole) (domain.Room, error)
}

type roomsService struct {
	roomsRepo repository.RoomsRepository
}

func NewRoomsService(roomsRepo repository.RoomsRepository) RoomsService {
	return &roomsService{
		roomsRepo: roomsRepo,
	}
}

func (s *roomsService) ListRooms(ctx context.Context) ([]domain.Room, error) {
	return s.roomsRepo.ListRooms(ctx)
}

func (s *roomsService) CreateRoom(ctx context.Context, input domain.RoomCreateInput, actorRole domain.UserRole) (domain.Room, error) {
	if actorRole != domain.RoleAdmin {
		return domain.Room{}, ErrForbidden
	}
	if strings.TrimSpace(input.Name) == "" {
		return domain.Room{}, ErrInvalidRoomInput
	}
	return s.roomsRepo.CreateRoom(ctx, input)
}
