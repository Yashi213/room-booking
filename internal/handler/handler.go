package handler

import (
	"time"

	"github.com/Yashi213/room-booking/internal/api"
	"github.com/Yashi213/room-booking/internal/auth"
	"github.com/Yashi213/room-booking/internal/service"
)

// Handler implements generated API interface.
type Handler struct {
	api.Unimplemented
	roomsService     service.RoomsService
	schedulesService service.SchedulesService
	slotsService     service.SlotsService
	bookingsService  service.BookingsService
	authService      service.AuthService
	jwtAccess        *auth.HS256Access
}

func New(
	roomsService service.RoomsService,
	schedulesService service.SchedulesService,
	slotsService service.SlotsService,
	bookingsService service.BookingsService,
	authService service.AuthService,
	jwtAccess *auth.HS256Access,
) *Handler {
	if jwtAccess == nil {
		jwtAccess = auth.NewHS256Access([]byte("supersecretkey"), 24*time.Hour)
	}
	return &Handler{
		roomsService:     roomsService,
		schedulesService: schedulesService,
		slotsService:     slotsService,
		bookingsService:  bookingsService,
		authService:      authService,
		jwtAccess:        jwtAccess,
	}
}

var (
	adminUserID = "11111111-1111-1111-1111-111111111111"
	userUserID  = "22222222-2222-2222-2222-222222222222"
)
