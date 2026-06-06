package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/trace/trace/pkg/auth"
)

// AuthHandler issues JWTs for the demo.
// In production this would validate credentials — for demo any userID works.
type AuthHandler struct {
	jwtSvc *auth.JWTService
}

func NewAuthHandler(jwtSvc *auth.JWTService) *AuthHandler {
	return &AuthHandler{jwtSvc: jwtSvc}
}

type tokenRequest struct {
	UserID string `json:"user_id"`
}

type tokenResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}

// POST /api/v1/auth/token
// Issues a JWT for the given user_id. Demo-only — no credential check.
func (h *AuthHandler) Token(w http.ResponseWriter, r *http.Request) {
	var req tokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID == "" {
		writeError(w, http.StatusBadRequest, "user_id required")
		return
	}
	token, err := h.jwtSvc.Issue(req.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not issue token")
		return
	}
	writeJSON(w, http.StatusOK, tokenResponse{Token: token, UserID: req.UserID})
}
