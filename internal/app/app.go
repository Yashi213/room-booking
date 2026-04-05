package app

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Yashi213/room-booking/internal/api"
	"github.com/Yashi213/room-booking/internal/auth"
	"github.com/Yashi213/room-booking/internal/conference"
	"github.com/Yashi213/room-booking/internal/handler"
	"github.com/Yashi213/room-booking/internal/repository"
	"github.com/Yashi213/room-booking/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	JWTSecret string

	ConferenceFail      bool
	ConferenceDelayMS   int
	ConferenceBadLink   string
	ConferenceTimeoutMS int
}

func Mount(_ context.Context, pool *pgxpool.Pool, r chi.Router, cfg Config) {
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "supersecretkey"
	}
	jwtTTL := jwtAccessTokenTTL()
	jwtAccess := auth.NewHS256Access([]byte(cfg.JWTSecret), jwtTTL)

	roomsRepo := repository.NewRoomsRepository(pool)
	schedulesRepo := repository.NewSchedulesRepository(pool)
	slotsRepo := repository.NewSlotsRepository(pool)
	bookingsRepo := repository.NewBookingsRepository(pool)
	usersRepo := repository.NewUsersRepository(pool)

	authService := service.NewAuthService(usersRepo, jwtAccess)
	roomsService := service.NewRoomsService(roomsRepo)
	slotsService := service.NewSlotsService(roomsRepo, slotsRepo, schedulesRepo)
	schedulesService := service.NewSchedulesService(roomsRepo, schedulesRepo, slotsService)
	confMock := &conference.Mock{
		ShouldFail: cfg.ConferenceFail,
		Delay:      time.Duration(cfg.ConferenceDelayMS) * time.Millisecond,
		BadLink:    cfg.ConferenceBadLink,
	}
	confTimeout := time.Duration(cfg.ConferenceTimeoutMS) * time.Millisecond
	if confTimeout <= 0 {
		confTimeout = 3 * time.Second
	}
	bookingsService := service.NewBookingsService(bookingsRepo, slotsRepo, confMock, confTimeout)

	h := handler.New(roomsService, schedulesService, slotsService, bookingsService, authService, jwtAccess)

	r.Get("/_info", h.Info)
	api.HandlerFromMux(h, r)
}

func NewRouter(ctx context.Context, pool *pgxpool.Pool, cfg Config) http.Handler {
	r := chi.NewRouter()
	Mount(ctx, pool, r, cfg)
	return r
}

func jwtAccessTokenTTL() time.Duration {
	const defaultHours = 24
	v := os.Getenv("JWT_TTL_HOURS")
	if v == "" {
		return defaultHours * time.Hour
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 || n > 720 {
		return defaultHours * time.Hour
	}
	return time.Duration(n) * time.Hour
}
