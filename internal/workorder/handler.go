package workorder

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"

	platformAuth "github.com/DashboardDivas/havenzsure-dashboard-backend/internal/platform/auth"

	"log"

	"github.com/DashboardDivas/havenzsure-dashboard-backend/internal/workorder/dto"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.ListWorkOrder)
	r.Post("/", h.CreateWorkOrder)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.GetWorkOrderByID)
		r.Get("/pdf", h.BuildWorkOrderPDF)
		r.Post("/email-report", h.EmailWorkOrderReport)
	})
}

// GET /workorders
func (h *Handler) ListWorkOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	actor, err := platformAuth.GetAuthUser(ctx)
	// log.Printf("[InjectUser] path=%s uid=%v", r.URL.Path, authUser.ID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !platformAuth.Can(platformAuth.RoleCode(actor.RoleCode), PermissionWorkOrderList) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	items, err := h.service.ListWorkOrder(ctx, actor.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
	log.Printf("[WO List] enter userID=%s", actor.ID)
}

// GET /workorders/{id}
func (h *Handler) GetWorkOrderByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authUser, err := platformAuth.GetAuthUser(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if !platformAuth.Can(platformAuth.RoleCode(authUser.RoleCode), PermissionWorkOrderDetail) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid UUID format", http.StatusBadRequest)
		return
	}

	wo, err := h.service.GetWorkOrderByID(ctx, authUser.ID, id)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wo)
}

// Post /workorders
func (h *Handler) CreateWorkOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	actor, err := platformAuth.GetAuthUser(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if !platformAuth.Can(platformAuth.RoleCode(actor.RoleCode), PermissionWorkOrderCreate) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var payload dto.IntakePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	wo, err := h.service.CreateWorkOrder(ctx, actor.ID, payload)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(wo)
}

// PUT /workorders/{code}/insurance

//Patch /workorders/{code}

func (h *Handler) BuildWorkOrderPDF(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser, err := platformAuth.GetAuthUser(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if !platformAuth.Can(platformAuth.RoleCode(authUser.RoleCode), PermissionWorkOrderDetail) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	idStr := chi.URLParam(r, "id")
	woID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid workorder id", http.StatusBadRequest)
		return
	}

	woDetail, err := h.service.GetWorkOrderByID(ctx, authUser.ID, woID)
	if err != nil {
		writeError(w, err)
		return
	}

	pdfBytes, err := BuildWorkOrderPDF(woDetail)
	if err != nil {
		http.Error(w, "Failed to build PDF", http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("%s.pdf", woDetail.Code)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdfBytes)
}

func (h *Handler) EmailWorkOrderReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser, err := platformAuth.GetAuthUser(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 和看详情 / 下 PDF 一样的权限
	if !platformAuth.Can(platformAuth.RoleCode(authUser.RoleCode), PermissionWorkOrderDetail) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	idStr := chi.URLParam(r, "id")
	woID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid workorder id", http.StatusBadRequest)
		return
	}

	woDetail, err := h.service.GetWorkOrderByID(ctx, authUser.ID, woID)
	if err != nil {
		writeError(w, err)
		return
	}

	pdfBytes, err := BuildWorkOrderPDF(woDetail)
	if err != nil {
		http.Error(w, "Failed to build PDF", http.StatusInternalServerError)
		return
	}

	sender, err := NewSMTPSenderFromEnv()
	if err != nil {
		http.Error(w, "SMTP not configured", http.StatusInternalServerError)
		return
	}

	err = sender.SendWorkOrderReport(
		ctx,
		woDetail.Customer.Email,
		woDetail.Customer.FullName,
		woDetail.Code,
		pdfBytes,
	)
	if err != nil {
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Work order report emailed successfully",
	})
}

// writeError maps errors to safe HTTP responses
// Note: Repository still returns detailed errors (fmt.Errorf),
// but this handler catches them all and only returns generic messages.
// TODO: Add mapPgError to repository.go for cleaner architecture
func writeError(w http.ResponseWriter, err error) {
	log.Printf("[WorkOrder ERROR] %v", err)

	// Note: pgx.ErrNoRows can be removed after repository maps it
	if errors.Is(err, ErrNotFound) || errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "work order not found", http.StatusNotFound)
		return
	}

	http.Error(w, "internal error", http.StatusInternalServerError)
}
