package handler

import "net/http"

// Info сервис жив.
// @Summary Служебный health
// @Tags System
// @Success 200
// @Router /_info [get]
func (h *Handler) Info(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
