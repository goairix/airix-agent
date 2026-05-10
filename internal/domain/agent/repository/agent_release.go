package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/domain/agent/model"
)

// ReleaseRepository AgentRelease 仓储接口
type ReleaseRepository interface {
	Save(ctx context.Context, release *model.AgentRelease) error
	FindByID(ctx context.Context, releaseID string) (*model.AgentRelease, error)
	FindActive(ctx context.Context, agentID uuid.UUID) (*model.AgentRelease, error)
	DeactivateAll(ctx context.Context, agentID uuid.UUID) error
	ListByAgent(ctx context.Context, agentID uuid.UUID, pagination Pagination) ([]model.AgentRelease, int64, error)
}
