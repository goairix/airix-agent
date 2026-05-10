package port

import (
	"context"

	"github.com/dysodeng/app/internal/domain/memory/model"
	sessionModel "github.com/dysodeng/app/internal/domain/session/model"
)

// MemoryExtractor 从会话中异步提取记忆（AfterAgent 钩子时机调用）
type MemoryExtractor interface {
	Extract(ctx context.Context, session *sessionModel.Session) ([]model.Memory, error)
}
