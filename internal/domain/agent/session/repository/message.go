package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/domain/agent/session/model"
)

// MessageRepository Message 仓储接口
type MessageRepository interface {
	Save(ctx context.Context, message *model.Message) error
	Update(ctx context.Context, message *model.Message) error
	FindByID(ctx context.Context, messageID uuid.UUID) (*model.Message, error)
	// ListBySession 按 sort_order 升序返回所有轮次
	ListBySession(ctx context.Context, sessionID uuid.UUID) ([]model.Message, error)
	// GetLatestN 取最近 n 轮，用于滑动窗口上下文组装
	GetLatestN(ctx context.Context, sessionID uuid.UUID, n int) ([]model.Message, error)
}
