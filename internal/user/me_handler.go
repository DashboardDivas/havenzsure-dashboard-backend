package user

import (
	"encoding/json"
	"net/http"

	"github.com/DashboardDivas/havenzsure-dashboard-backend/internal/platform/auth"
	"github.com/go-chi/chi/v5"
)

// MeHandler for handling /me routes
type MeHandler struct {
	svc MeService
}

func NewMeHandler(svc MeService) *MeHandler {
	return &MeHandler{svc: svc}
}

// RegisterRoutes registers routes under the /me route group
func (h *MeHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.getCurrentUser)
	r.Put("/", h.updateProfile)
}

// GET /me
func (h *MeHandler) getCurrentUser(w http.ResponseWriter, r *http.Request) {
	// 1) Get GCIP / Firebase auth user from context
	authUser, err := auth.GetAuthUser(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	// 2) Use externalID to query DB user
	user, err := h.svc.GetCurrentUser(r.Context(), authUser.ExternalID)
	if err != nil {
		writeError(w, err)
		return
	}

	// 3) Return the current user's UserResponse
	writeJSON(w, http.StatusOK, user.ToResponse())
}

// PUT /me
func (h *MeHandler) updateProfile(w http.ResponseWriter, r *http.Request) {
	var in UpdateMeProfileInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, ErrInvalidInput)
		return
	}
	// 1) Get GCIP / Firebase auth user from context
	authUser, err := auth.GetAuthUser(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	// 2) Delegate to MeService.UpdateProfile
	updatedUser, err := h.svc.UpdateProfile(r.Context(), authUser.ExternalID, &in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updatedUser.ToResponse())
}
