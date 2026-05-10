package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/domain/agent/model"
)

// Pagination 分页参数
type Pagination struct {
	Page     int
	PageSize int
}

// Repository Agent 仓储接口
type Repository interface {
	Save(ctx context.Context, agent *model.Agent) error
	FindByID(ctx context.Context, agentID uuid.UUID) (*model.Agent, error)
	FindByWorkspace(ctx context.Context, workspaceID uuid.UUID, pagination Pagination) ([]model.Agent, int64, error)
	Delete(ctx context.Context, agentID uuid.UUID) error
}
