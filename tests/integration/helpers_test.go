package integration

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func integrationDBURL(t *testing.T) string {
	t.Helper()
	u := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if u == "" {
		t.Skip("интеграционные тесты: задайте DATABASE_URL (Postgres)")
	}
	return u
}

func connectPool(t *testing.T, url string) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), url)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return pool
}

// deleteTestRoom удаляет комнату и через ON DELETE CASCADE — расписание, слоты и брони.
func deleteTestRoom(tb testing.TB, pool *pgxpool.Pool, roomID string) {
	tb.Helper()
	if roomID == "" {
		return
	}
	ctx := context.Background()
	if _, err := pool.Exec(ctx, `DELETE FROM rooms WHERE id = $1`, roomID); err != nil {
		tb.Logf("integration: удаление тестовой комнаты: %v", err)
	}
}

func apiDayOfWeek(d time.Time) int {
	switch d.Weekday() {
	case time.Sunday:
		return 7
	default:
		return int(d.Weekday())
	}
}
