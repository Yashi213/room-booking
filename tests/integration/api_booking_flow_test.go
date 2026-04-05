package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Yashi213/room-booking/internal/app"
	"github.com/stretchr/testify/require"
)

func TestIntegration_RoomScheduleBooking(t *testing.T) {
	dbURL := integrationDBURL(t)
	pool := connectPool(t, dbURL)

	h := app.NewRouter(context.Background(), pool, app.Config{JWTSecret: "integration-test-jwt"})
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	base := srv.URL

	adminTok := dummyLogin(t, base, "admin")
	roomID := createRoom(t, base, adminTok, "Integration Room A")
	t.Cleanup(func() { deleteTestRoom(t, pool, roomID) })

	testDay := time.Now().UTC().Add(48 * time.Hour).Truncate(24 * time.Hour)
	dow := apiDayOfWeek(testDay)
	createSchedule(t, base, adminTok, roomID, dow, "09:00", "18:00")

	userTok := dummyLogin(t, base, "user")
	slots := listSlots(t, base, userTok, roomID, testDay.Format("2006-01-02"))
	require.NotEmpty(t, slots, "слоты должны появиться в запланированный день")

	slotObj := slots[0].(map[string]any)
	slotID, _ := slotObj["id"].(string)
	require.NotEmpty(t, slotID)
	bookingID := createBooking(t, base, userTok, slotID)
	require.NotEmpty(t, bookingID)
}

func TestIntegration_CancelBooking_Idempotent(t *testing.T) {
	dbURL := integrationDBURL(t)
	pool := connectPool(t, dbURL)

	h := app.NewRouter(context.Background(), pool, app.Config{JWTSecret: "integration-test-jwt"})
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	base := srv.URL

	adminTok := dummyLogin(t, base, "admin")
	roomID := createRoom(t, base, adminTok, "Integration Room B")
	t.Cleanup(func() { deleteTestRoom(t, pool, roomID) })

	testDay := time.Now().UTC().Add(72 * time.Hour).Truncate(24 * time.Hour)
	dow := apiDayOfWeek(testDay)
	createSchedule(t, base, adminTok, roomID, dow, "10:00", "12:00")

	userTok := dummyLogin(t, base, "user")
	slots := listSlots(t, base, userTok, roomID, testDay.Format("2006-01-02"))
	require.NotEmpty(t, slots)
	slotObj := slots[0].(map[string]any)
	slotID, _ := slotObj["id"].(string)
	bookingID := createBooking(t, base, userTok, slotID)

	code1 := postCancel(t, base, userTok, bookingID)
	code2 := postCancel(t, base, userTok, bookingID)
	require.Equal(t, http.StatusOK, code1)
	require.Equal(t, http.StatusOK, code2, "повторная отмена — тоже 200")
}

func dummyLogin(t *testing.T, base, role string) string {
	t.Helper()
	body := map[string]string{"role": role}
	b, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, base+"/dummyLogin", bytes.NewReader(b))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(raw))

	var out struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(raw, &out))
	require.NotEmpty(t, out.Token)
	return out.Token
}

func createRoom(t *testing.T, base, bearer, name string) string {
	t.Helper()
	body := map[string]string{"name": name}
	b, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, base+"/rooms/create", bytes.NewReader(b))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bearer)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusCreated, resp.StatusCode, string(raw))

	var out struct {
		Room map[string]any `json:"room"`
	}
	require.NoError(t, json.Unmarshal(raw, &out))
	id, _ := out.Room["id"].(string)
	require.NotEmpty(t, id)
	return id
}

func createSchedule(t *testing.T, base, bearer, roomID string, dayOfWeek int, start, end string) {
	t.Helper()
	body := map[string]any{
		"daysOfWeek": []int{dayOfWeek},
		"startTime":  start,
		"endTime":    end,
	}
	b, _ := json.Marshal(body)
	u := fmt.Sprintf("%s/rooms/%s/schedule/create", base, roomID)
	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(b))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bearer)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusCreated, resp.StatusCode, string(raw))
}

func listSlots(t *testing.T, base, bearer, roomID, date string) []any {
	t.Helper()
	u := fmt.Sprintf("%s/rooms/%s/slots/list?date=%s", base, roomID, date)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+bearer)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(raw))

	var out struct {
		Slots []any `json:"slots"`
	}
	require.NoError(t, json.Unmarshal(raw, &out))
	return out.Slots
}

func createBooking(t *testing.T, base, bearer, slotID string) string {
	t.Helper()
	body := map[string]string{"slotId": slotID}
	b, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, base+"/bookings/create", bytes.NewReader(b))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bearer)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusCreated, resp.StatusCode, string(raw))

	var out struct {
		Booking map[string]any `json:"booking"`
	}
	require.NoError(t, json.Unmarshal(raw, &out))
	id, _ := out.Booking["id"].(string)
	return id
}

func postCancel(t *testing.T, base, bearer, bookingID string) int {
	t.Helper()
	u := fmt.Sprintf("%s/bookings/%s/cancel", base, bookingID)
	req, err := http.NewRequest(http.MethodPost, u, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+bearer)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	return resp.StatusCode
}
