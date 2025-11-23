package workorder

import (
	"encoding/json"
	"net/http"

	"github.com/DashboardDivas/havenzsure-dashboard-backend/internal/workorder/dto"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.ListWorkOrder)
	r.Get("/{code}", h.GetWorkOrderByCode)
	r.Post("/", h.CreateWorkOrder)
	// r.Put("/{code}/insurance", h.UpsertInsurance)

}

// GET /workorders
func (h *Handler) ListWorkOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	items, err := h.service.ListWorkOrder(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// GET /workorders/{code}
func (h *Handler) GetWorkOrderByCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := chi.URLParam(r, "code")

	wo, err := h.service.GetWorkOrderByCode(ctx, code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wo)
}

// Post /workorders
func (h *Handler) CreateWorkOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var payload dto.IntakePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	wo, err := h.service.CreateWorkOrder(ctx, payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(wo)
}

// PUT /workorders/{code}/insurance

//Patch /workorders/{code}
