package unit

import (
	"context"
	"testing"

	"github.com/Yashi213/room-booking/internal/domain"
	"github.com/Yashi213/room-booking/internal/mocks"
	"github.com/Yashi213/room-booking/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRoomsService_CreateRoom_ForbiddenNonAdmin(t *testing.T) {
	svc := service.NewRoomsService(nil)
	_, err := svc.CreateRoom(context.Background(), domain.RoomCreateInput{Name: "A"}, domain.RoleUser)
	require.ErrorIs(t, err, service.ErrForbidden)
}

func TestRoomsService_CreateRoom_EmptyName(t *testing.T) {
	svc := service.NewRoomsService(nil)
	_, err := svc.CreateRoom(context.Background(), domain.RoomCreateInput{Name: "  "}, domain.RoleAdmin)
	require.ErrorIs(t, err, service.ErrInvalidRoomInput)
}

func TestRoomsService_CreateRoom_Success(t *testing.T) {
	repo := mocks.NewRoomsRepository(t)
	room := domain.Room{ID: uuid.New(), Name: "Alpha"}
	repo.EXPECT().CreateRoom(mock.Anything, mock.MatchedBy(func(in domain.RoomCreateInput) bool {
		return in.Name == "Alpha"
	})).Return(room, nil)

	svc := service.NewRoomsService(repo)
	out, err := svc.CreateRoom(context.Background(), domain.RoomCreateInput{Name: "Alpha"}, domain.RoleAdmin)
	require.NoError(t, err)
	require.Equal(t, "Alpha", out.Name)
}

func TestRoomsService_ListRooms(t *testing.T) {
	repo := mocks.NewRoomsRepository(t)
	want := []domain.Room{{ID: uuid.New(), Name: "R1"}}
	repo.EXPECT().ListRooms(mock.Anything).Return(want, nil)

	svc := service.NewRoomsService(repo)
	out, err := svc.ListRooms(context.Background())
	require.NoError(t, err)
	require.Equal(t, want, out)
}
