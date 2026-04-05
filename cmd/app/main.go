// @title Room Booking Service
// @version 1.0.0
// @description Сервис бронирования переговорок. Защищённые методы: заголовок Authorization: Bearer {token} (токен из POST /dummyLogin).
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/Yashi213/room-booking/docs"
	"github.com/Yashi213/room-booking/internal/app"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/roombooking?sslmode=disable"
	}

	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	mockFail, _ := strconv.ParseBool(strings.TrimSpace(os.Getenv("MOCK_CONFERENCE_FAIL")))
	delayMS, _ := strconv.Atoi(strings.TrimSpace(os.Getenv("MOCK_CONFERENCE_DELAY_MS")))
	confTimeoutMS, _ := strconv.Atoi(strings.TrimSpace(os.Getenv("CONFERENCE_SERVICE_TIMEOUT_MS")))

	r := chi.NewRouter()
	app.Mount(context.Background(), dbPool, r, app.Config{
		JWTSecret:           os.Getenv("JWT_SECRET"),
		ConferenceFail:      mockFail,
		ConferenceDelayMS:   delayMS,
		ConferenceBadLink:   os.Getenv("MOCK_CONFERENCE_BAD_LINK"),
		ConferenceTimeoutMS: confTimeoutMS,
	})
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	const addr = ":8080"
	log.Printf("server listening on %s (swagger: /swagger/index.html)", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
