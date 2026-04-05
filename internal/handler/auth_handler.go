package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Yashi213/room-booking/internal/api"
	"github.com/Yashi213/room-booking/internal/mapper"
	"github.com/google/uuid"
)

func (h *Handler) PostDummyLogin(w http.ResponseWriter, r *http.Request) {
	var req api.PostDummyLoginJSONBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	role := string(req.Role)
	var userID uuid.UUID
	var err error
	switch role {
	case "admin":
		userID, err = uuid.Parse(adminUserID)
	case "user":
		userID, err = uuid.Parse(userUserID)
	default:
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "role must be admin or user")
		return
	}
	if err != nil {
		writeInternalError(w)
		return
	}

	signed, err := h.jwtAccess.SignAccessToken(userID, role)
	if err != nil {
		writeInternalError(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(api.Token{Token: signed})
}

func (h *Handler) PostRegister(w http.ResponseWriter, r *http.Request) {
	var body api.PostRegisterJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}
	if !body.Role.Valid() {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid role")
		return
	}

	u, err := h.authService.Register(r.Context(), mapper.ToUserRegisterInput(body))
	if err != nil {
		if mappedErr, ok := mapServiceError(err); ok {
			writeError(w, mappedErr.status, mappedErr.code, mappedErr.message)
			return
		}
		writeInternalError(w)
		return
	}

	dto := mapper.ToUserDTO(u)
	resp := api.PostRegister201JSONResponse{User: &dto}
	if err := resp.VisitPostRegisterResponse(w); err != nil {
		writeInternalError(w)
	}
}

func (h *Handler) PostLogin(w http.ResponseWriter, r *http.Request) {
	var body api.PostLoginJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	token, err := h.authService.Login(r.Context(), mapper.ToUserLoginInput(body))
	if err != nil {
		if mappedErr, ok := mapServiceError(err); ok {
			writeError(w, mappedErr.status, mappedErr.code, mappedErr.message)
			return
		}
		writeInternalError(w)
		return
	}

	resp := api.PostLogin200JSONResponse{Token: token}
	if err := resp.VisitPostLoginResponse(w); err != nil {
		writeInternalError(w)
	}
}
