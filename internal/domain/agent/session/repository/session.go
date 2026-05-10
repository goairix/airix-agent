package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/domain/agent/session/model"
)

// Pagination 分页参数
type Pagination struct {
	Page     int
	PageSize int
}

// SessionRepository Session 仓储接口
type SessionRepository interface {
	Save(ctx context.Context, session *model.Session) error
	FindByID(ctx context.Context, sessionID uuid.UUID) (*model.Session, error)
	ListByAgent(ctx context.Context, agentID uuid.UUID, pagination Pagination) ([]model.Session, int64, error)
}
