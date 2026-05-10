package port

import (
	"context"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/domain/memory/model"
)

// GlobalMemoryStore 全局记忆存取端口（可插拔驱动）
type GlobalMemoryStore interface {
	Upsert(ctx context.Context, entry *model.Memory) error
	LoadAll(ctx context.Context, workspaceID, userID uuid.UUID) ([]model.Memory, error)
	Search(ctx context.Context, workspaceID, userID uuid.UUID, query string) ([]model.Memory, error)
}
