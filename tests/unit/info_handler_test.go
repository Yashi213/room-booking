package unit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yashi213/room-booking/internal/handler"
	"github.com/stretchr/testify/require"
)

func TestHandler_Info_OK(t *testing.T) {
	h := handler.New(nil, nil, nil, nil, nil, nil)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/_info", nil)
	h.Info(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
}
