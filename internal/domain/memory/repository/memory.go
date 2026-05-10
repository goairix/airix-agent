package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/domain/memory/model"
)

// MemoryRepository Memory 数据库仓储接口
type MemoryRepository interface {
	Save(ctx context.Context, memory *model.Memory) error
	ListByUserAndDate(ctx context.Context, workspaceID, userID uuid.UUID, date time.Time) ([]model.Memory, error)
	// ListByAgentUser 查询某 agent+用户的所有会话记忆（session 类型）
	ListByAgentUser(ctx context.Context, workspaceID, agentID, userID uuid.UUID) ([]model.Memory, error)
	// DeleteByAgentUser 清除指定 agent+用户的所有会话记忆（session 类型）
	DeleteByAgentUser(ctx context.Context, workspaceID, agentID, userID uuid.UUID) error
}
