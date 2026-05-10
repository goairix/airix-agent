package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/domain/session/model"
)

// MessageItemRepository MessageItem 仓储接口
type MessageItemRepository interface {
	BatchSave(ctx context.Context, items []*model.MessageItem) error
	// ListByMessage 取该轮次所有步骤（内存排序由调用方负责）
	ListByMessage(ctx context.Context, messageID uuid.UUID) ([]*model.MessageItem, error)
}
