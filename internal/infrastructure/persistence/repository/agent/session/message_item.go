package session

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	sessionModel "github.com/dysodeng/app/internal/domain/agent/session/model"
	"github.com/dysodeng/app/internal/domain/agent/session/repository"
	"github.com/dysodeng/app/internal/domain/agent/session/valueobject"
	sessionEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/agent/session"
	"github.com/dysodeng/app/internal/infrastructure/persistence/transactions"
	pkgModel "github.com/dysodeng/app/internal/infrastructure/pkg/model"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

type messageItemRepository struct {
	baseTraceSpanName string
	txManager         transactions.TransactionManager
}

func NewMessageItemRepository(txManager transactions.TransactionManager) repository.MessageItemRepository {
	return &messageItemRepository{
		baseTraceSpanName: "infrastructure.persistence.repository.session.MessageItemRepository",
		txManager:         txManager,
	}
}

func (repo *messageItemRepository) BatchSave(ctx context.Context, items []*sessionModel.MessageItem) error {
	if len(items) == 0 {
		return nil
	}
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".BatchSave")
	defer span.End()
	entities := make([]*sessionEntity.MessageItem, 0, len(items))
	for _, item := range items {
		e, err := repo.toEntity(item)
		if err != nil {
			return err
		}
		entities = append(entities, e)
	}
	return repo.txManager.GetTx(spanCtx).Create(&entities).Error
}

func (repo *messageItemRepository) ListByMessage(ctx context.Context, messageID uuid.UUID) ([]*sessionModel.MessageItem, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".ListByMessage")
	defer span.End()
	var entities []sessionEntity.MessageItem
	if err := repo.txManager.GetTx(spanCtx).
		Where("message_id = ?", messageID).Find(&entities).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	items := make([]*sessionModel.MessageItem, 0, len(entities))
	for _, e := range entities {
		item, err := repo.fromEntity(&e)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (repo *messageItemRepository) toEntity(item *sessionModel.MessageItem) (*sessionEntity.MessageItem, error) {
	contentJSON, err := json.Marshal(item.Content)
	if err != nil {
		return nil, err
	}
	return &sessionEntity.MessageItem{
		DistributedPrimaryKeyID: pkgModel.DistributedPrimaryKeyID{ID: item.ID},
		MessageID:               item.MessageID,
		SessionID:               item.SessionID,
		SortOrder:               item.SortOrder,
		ItemType:                item.ItemType.Uint8(),
		IsFinal:                 item.IsFinal,
		Content:                 string(contentJSON),
		InputTokens:             item.InputTokens,
		OutputTokens:            item.OutputTokens,
		LatencyMs:               item.LatencyMs,
		CreatedAt:               item.CreatedAt,
	}, nil
}

func (repo *messageItemRepository) fromEntity(e *sessionEntity.MessageItem) (*sessionModel.MessageItem, error) {
	var content sessionModel.MessageItemContent
	if err := json.Unmarshal([]byte(e.Content), &content); err != nil {
		return nil, err
	}
	return &sessionModel.MessageItem{
		ID:           e.ID,
		MessageID:    e.MessageID,
		SessionID:    e.SessionID,
		SortOrder:    e.SortOrder,
		ItemType:     valueobject.MessageItemType(e.ItemType),
		IsFinal:      e.IsFinal,
		Content:      content,
		InputTokens:  e.InputTokens,
		OutputTokens: e.OutputTokens,
		LatencyMs:    e.LatencyMs,
		CreatedAt:    e.CreatedAt,
	}, nil
}
