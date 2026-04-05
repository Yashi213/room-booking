package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"

	"github.com/Yashi213/room-booking/internal/repository"
	"github.com/Yashi213/room-booking/internal/service"
)

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/roombooking?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	roomsRepo := repository.NewRoomsRepository(pool)
	schedRepo := repository.NewSchedulesRepository(pool)
	slotsRepo := repository.NewSlotsRepository(pool)
	slotsSvc := service.NewSlotsService(roomsRepo, slotsRepo, schedRepo)

	run := func() {
		ctx := context.Background()
		if err := slotsSvc.RefillRollingWindowSlots(ctx); err != nil {
			log.Printf("slotworker: refill: %v", err)
			return
		}
		log.Println("slotworker: rolling window slots ensured")
	}

	run()

	const cronExpr = "0 3 * * *"
	c := cron.New(cron.WithLocation(time.UTC))
	if _, err := c.AddFunc(cronExpr, run); err != nil {
		log.Fatalf("slotworker: cron %q: %v", cronExpr, err)
	}
	c.Start()
	log.Printf("slotworker: cron %q (UTC ежедневно 03:00)", cronExpr)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	<-c.Stop().Done()
	log.Println("slotworker: shutdown")
}
