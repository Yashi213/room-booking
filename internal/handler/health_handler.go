package handler

import "net/http"

func (h *Handler) Info(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
