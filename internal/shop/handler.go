// Citation: Chatgpt prompts: can you help me structure my handler.go based on repository.go, model.go, errors.go and service.go?
// internal/shop/handler.go
package shop

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Handler wires HTTP endpoints to the ShopService.
// It contains no persistence details—only request parsing, calling the service,
// and mapping domain errors to HTTP responses.
type Handler struct {
	svc ShopService
}

// NewHandler constructs a shop HTTP handler that depends on a ShopService.
func NewHandler(svc ShopService) *Handler { return &Handler{svc: svc} }

// RegisterRoutes mounts the /shops endpoints on the given router.
// Endpoints (code-centric resource paths):
//
//	POST /shops          -> create a shop
//	GET  /shops          -> list shops (supports ?limit=&offset=)
//	GET  /shops/{code}   -> fetch one shop by code
//	PUT  /shops/{code}   -> update a shop identified by code
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.create)
	r.Get("/", h.list)
	r.Get("/{code}", h.getByCode)
	r.Put("/{code}", h.updateByCode)
}

// create handles POST /shops.
// - Decodes the JSON payload into a Shop.
// - Delegates creation to the service (which fills fields like ID).
// - Returns 201 with Location header pointing to /shops/{code}.
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var s Shop
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		httpError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.svc.CreateShop(r.Context(), &s); err != nil {
		writeError(w, err)
		return
	}

	// Since users navigate by code, expose the resource URL by code.
	w.Header().Set("Location", fmt.Sprintf("/shops/%s", s.Code))
	writeJSON(w, http.StatusCreated, &s)
}

// getByCode handles GET /shops/{code}.
// - Extracts the code from the path.
// - Retrieves the shop from the service.
// - Returns 200 with the shop JSON.
func (h *Handler) getByCode(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSpace(chi.URLParam(r, "code"))
	out, err := h.svc.GetShopByCode(r.Context(), code)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// list handles GET /shops?limit=&offset=.
// - Parses pagination params with safe defaults.
// - Delegates listing to the service.
// - Returns 200 with an array of shops.
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	limit := atoiDefault(r.URL.Query().Get("limit"), 50)
	offset := atoiDefault(r.URL.Query().Get("offset"), 0)
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	out, err := h.svc.ListShops(r.Context(), limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// updateByCode handles PUT /shops/{code}.
// Because the write API uses ID in the repository, we:
//  1. Look up the existing record by code to get its ID.
//  2. Decode the request body.
//  3. Set in.ID to the current record's ID (so we update the correct row).
//  4. Call UpdateShop and return the updated record.
func (h *Handler) updateByCode(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSpace(chi.URLParam(r, "code"))

	current, err := h.svc.GetShopByCode(r.Context(), code)
	if err != nil {
		writeError(w, err)
		return
	}

	var in Shop
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpError(w, http.StatusBadRequest, "invalid json")
		return
	}

	// Ensure we update the resource identified by the path.
	in.ID = current.ID

	out, err := h.svc.UpdateShop(r.Context(), &in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// ---- helpers ----

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

// writeJSON marshals v as JSON and writes it with the provided status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// httpError is a small helper for writing an error envelope with a status/code/message shape.
func httpError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{
		"code":    http.StatusText(status),
		"message": msg,
	})
}

// writeError maps domain-level errors to HTTP status codes.
// - ErrInvalidInput → 400
// - ErrConflict     → 409
// - ErrNotFound     → 404
// - others          → 500
func writeError(w http.ResponseWriter, err error) {
	// Log the error for server-side diagnostics
	log.Printf("[ERROR] %v", err)

	// Map domain errors to HTTP status codes
	switch {
	case errors.Is(err, ErrInvalidInput):
		httpError(w, http.StatusBadRequest, "invalid input")
	case errors.Is(err, ErrConflict):
		httpError(w, http.StatusConflict, "conflict")
	case errors.Is(err, ErrNotFound):
		httpError(w, http.StatusNotFound, "not found")
	default:
		httpError(w, http.StatusInternalServerError, "internal error")
	}
}
