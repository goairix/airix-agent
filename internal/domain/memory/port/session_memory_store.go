package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/domain/memory/model"
)

// SearchOptions 会话记忆检索选项
type SearchOptions struct {
	TopK        int
	WorkspaceID uuid.UUID
	AgentID     uuid.UUID // 会话记忆按 Agent+User 区分
	UserID      uuid.UUID
}

// SessionMemoryStore 会话记忆存取端口（可插拔驱动）
type SessionMemoryStore interface {
	Save(ctx context.Context, entry *model.Memory) error
	Search(ctx context.Context, query string, opts SearchOptions) ([]model.Memory, error)
	ListByDate(ctx context.Context, workspaceID, userID uuid.UUID, date time.Time) ([]model.Memory, error)
}
