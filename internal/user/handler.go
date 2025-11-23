package user

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

/* -------------------- Handler Struct -------------------- */

type Handler struct {
	svc UserService
}

func NewHandler(svc UserService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.create)
	r.Get("/", h.list)
	r.Get("/{id}", h.getByID)
	r.Put("/{id}", h.update)
	r.Put("/{id}/deactivate", h.deactivate)
	r.Put("/{id}/reactivate", h.reactivate)
	r.Post("/{id}/resend-password-link", h.resendPasswordSetupLink)
}

/* -------------------- CRUD Handlers -------------------- */

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var in CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, ErrInvalidInput)
		return
	}

	user, err := h.svc.CreateUser(r.Context(), &in)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, user.ToResponse())
}

func (h *Handler) getByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, ErrInvalidInput)
		return
	}

	user, err := h.svc.GetUserByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, user.ToResponse())
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	limit := atoiDefault(r.URL.Query().Get("limit"), 50)
	offset := atoiDefault(r.URL.Query().Get("offset"), 0)
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	users, err := h.svc.ListUsers(r.Context(), limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}

	response := make([]*UserResponse, len(users))
	for i, u := range users {
		response[i] = u.ToResponse()
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(chi.URLParam(r, "id"))
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, ErrInvalidInput)
		return
	}

	var in UpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, ErrInvalidInput)
		return
	}

	user, err := h.svc.UpdateUser(r.Context(), id, &in)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, user.ToResponse())
}

func (h *Handler) deactivate(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(chi.URLParam(r, "id"))
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, ErrInvalidInput)
		return
	}

	if err := h.svc.DeactivateUser(r.Context(), id, nil); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) reactivate(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(chi.URLParam(r, "id"))
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, ErrInvalidInput)
		return
	}

	if err := h.svc.ReactivateUser(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// resendPasswordSetupLink resends password setup link to user
// POST /users/{id}/resend-password-link
func (h *Handler) resendPasswordSetupLink(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(chi.URLParam(r, "id"))
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, ErrInvalidInput)
		return
	}

	if err := h.svc.ResendPasswordSetupLink(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}

	// Return success message
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Password setup link sent successfully",
	})
}

/* -------------------- Helpers -------------------- */

// atoiDefault parses an int or returns a default value if parsing fails or s is empty.
func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// httpError outputs a uniform JSON error structure like shop handler.
func httpError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// writeError classifies known domain errors and delegates to httpError.
func writeError(w http.ResponseWriter, err error) {
	// Log the error for server-side diagnostics
	log.Printf("[ERROR] %v", err)

	// Map domain errors to HTTP status codes
	switch {

	case errors.Is(err, ErrInvalidInput):
		httpError(w, http.StatusBadRequest, err.Error())

	case errors.Is(err, ErrNotFound):
		httpError(w, http.StatusNotFound, err.Error())

	case errors.Is(err, ErrConflict):
		httpError(w, http.StatusConflict, err.Error())

	default:
		httpError(w, http.StatusInternalServerError, "internal server error")
	}
}
