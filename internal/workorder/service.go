package workorder

import (
	"context"
	"strings"

	"github.com/DashboardDivas/havenzsure-dashboard-backend/internal/workorder/dto"
)

type Service interface {
	ListWorkOrder(ctx context.Context) ([]dto.WorkOrderListItem, error)
	GetWorkOrderByCode(ctx context.Context, code string) (dto.WorkOrderDetail, error)
	CreateWorkOrder(ctx context.Context, payload dto.IntakePayload) (dto.WorkOrderDetail, error)
	//UpsertInsurance(ctx context.Context, workOrderID string, payload dto.InsuranceIntake) (dto.WorkOrderDetail, error)
	//EditIntake(ctx context.Context, code string, payload dto.IntakeEditPayload) (dto.WorkOrderDetail, error)
}

type service struct {
	repo Repository
}

var _ Service = (*service)(nil)

func NewService(r Repository) Service {
	return &service{repo: r}
}

func (s *service) ListWorkOrder(ctx context.Context) ([]dto.WorkOrderListItem, error) {
	return s.repo.ListWorkOrder(ctx)
}
func (s *service) GetWorkOrderByCode(ctx context.Context, code string) (dto.WorkOrderDetail, error) {
	return s.repo.GetWorkOrderByCode(ctx, code)
}
func (s *service) CreateWorkOrder(ctx context.Context, payload dto.IntakePayload) (dto.WorkOrderDetail, error) {
	return s.repo.CreateWorkOrder(ctx, payload)
}

// func (s *service) UpsertInsurance(ctx context.Context, workOrderID string, payload dto.InsuranceIntake) error {
// 	return s.repo.UpsertInsurance(ctx, workOrderID, payload)
// }

// func (s *service) EditIntake(ctx context.Context, code string, payload dto.IntakeEditPayload) (dto.WorkOrderDetail, error) {
// 	return s.repo.EditIntake(ctx, code, payload)
// }

// For InsuranceIntake, check if strings from frontend are empty string
// If so, trim it and set to nil
func nullIfEmpty(s string) *string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
