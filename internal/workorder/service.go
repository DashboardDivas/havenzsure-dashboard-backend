package workorder

import (
	"context"

	"github.com/DashboardDivas/havenzsure-dashboard-backend/internal/workorder/dto"
)

type Service interface {
	List(ctx context.Context) ([]dto.WorkOrderListItem, error)
	GetByCode(ctx context.Context, code string) (dto.WorkOrderDetail, error)
}

type service struct {
	repo Repository
}

var _ Service = (*service)(nil)

func NewService(r Repository) Service {
	return &service{repo: r}
}

func (s *service) List(ctx context.Context) ([]dto.WorkOrderListItem, error) {
	return s.repo.List(ctx)
}
func (s *service) GetByCode(ctx context.Context, code string) (dto.WorkOrderDetail, error) {
	return s.repo.GetByCode(ctx, code)
}
