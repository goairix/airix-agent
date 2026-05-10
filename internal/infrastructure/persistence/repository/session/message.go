package session

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	sessionModel "github.com/dysodeng/app/internal/domain/session/model"
	"github.com/dysodeng/app/internal/domain/session/repository"
	"github.com/dysodeng/app/internal/domain/session/valueobject"
	sessionEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/session"
	"github.com/dysodeng/app/internal/infrastructure/persistence/transactions"
	pkgModel "github.com/dysodeng/app/internal/infrastructure/pkg/model"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

type messageRepository struct {
	baseTraceSpanName string
	txManager         transactions.TransactionManager
}

func NewMessageRepository(txManager transactions.TransactionManager) repository.MessageRepository {
	return &messageRepository{
		baseTraceSpanName: "infrastructure.persistence.repository.session.MessageRepository",
		txManager:         txManager,
	}
}

func (repo *messageRepository) Save(ctx context.Context, m *sessionModel.Message) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Save")
	defer span.End()
	entity, err := repo.toEntity(m)
	if err != nil {
		return err
	}
	tx := repo.txManager.GetTx(spanCtx)
	if err = tx.Create(entity).Error; err != nil {
		return err
	}
	m.ID = entity.ID
	return nil
}

func (repo *messageRepository) Update(ctx context.Context, m *sessionModel.Message) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Update")
	defer span.End()
	entity, err := repo.toEntity(m)
	if err != nil {
		return err
	}
	return repo.txManager.GetTx(spanCtx).Where("id = ?", entity.ID).Updates(entity).Error
}

func (repo *messageRepository) FindByID(ctx context.Context, messageID uuid.UUID) (*sessionModel.Message, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindByID")
	defer span.End()
	var entity sessionEntity.Message
	if err := repo.txManager.GetTx(spanCtx).Where("id = ?", messageID).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return repo.fromEntity(&entity)
}

func (repo *messageRepository) ListBySession(ctx context.Context, sessionID uuid.UUID) ([]sessionModel.Message, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".ListBySession")
	defer span.End()
	var entities []sessionEntity.Message
	if err := repo.txManager.GetTx(spanCtx).
		Where("session_id = ?", sessionID).
		Order("sort_order ASC").Find(&entities).Error; err != nil {
		return nil, err
	}
	return repo.fromEntities(entities)
}

func (repo *messageRepository) GetLatestN(ctx context.Context, sessionID uuid.UUID, n int) ([]sessionModel.Message, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".GetLatestN")
	defer span.End()
	var entities []sessionEntity.Message
	if err := repo.txManager.GetTx(spanCtx).
		Where("session_id = ?", sessionID).
		Order("sort_order DESC").Limit(n).Find(&entities).Error; err != nil {
		return nil, err
	}
	// 反转使其升序
	for i, j := 0, len(entities)-1; i < j; i, j = i+1, j-1 {
		entities[i], entities[j] = entities[j], entities[i]
	}
	return repo.fromEntities(entities)
}

func (repo *messageRepository) fromEntities(entities []sessionEntity.Message) ([]sessionModel.Message, error) {
	messages := make([]sessionModel.Message, 0, len(entities))
	for _, e := range entities {
		m, err := repo.fromEntity(&e)
		if err != nil {
			return nil, err
		}
		messages = append(messages, *m)
	}
	return messages, nil
}

func (repo *messageRepository) toEntity(m *sessionModel.Message) (*sessionEntity.Message, error) {
	var agentInputJSON string
	if m.AgentInput != nil {
		b, err := json.Marshal(m.AgentInput)
		if err != nil {
			return nil, err
		}
		agentInputJSON = string(b)
	}
	return &sessionEntity.Message{
		DistributedPrimaryKeyID: pkgModel.DistributedPrimaryKeyID{ID: m.ID},
		SessionID:               m.SessionID,
		WorkspaceID:             m.WorkspaceID,
		AgentID:                 m.AgentID,
		SortOrder:               m.SortOrder,
		Query:                   m.Query,
		AgentInput:              agentInputJSON,
		Status:                  m.Status.Uint8(),
		TotalTokens:             m.TotalTokens,
		InputTokens:             m.InputTokens,
		OutputTokens:            m.OutputTokens,
		CachedTokens:            m.CachedTokens,
		ExecutionTimeMs:         m.ExecutionTimeMs,
		FirstTokenLatencyMs:     m.FirstTokenLatencyMs,
		CreatedAt:               m.CreatedAt,
		CompletedAt:             m.CompletedAt,
	}, nil
}

func (repo *messageRepository) fromEntity(e *sessionEntity.Message) (*sessionModel.Message, error) {
	m := &sessionModel.Message{
		ID:                  e.ID,
		SessionID:           e.SessionID,
		WorkspaceID:         e.WorkspaceID,
		AgentID:             e.AgentID,
		SortOrder:           e.SortOrder,
		Query:               e.Query,
		Status:              valueobject.MessageStatus(e.Status),
		TotalTokens:         e.TotalTokens,
		InputTokens:         e.InputTokens,
		OutputTokens:        e.OutputTokens,
		CachedTokens:        e.CachedTokens,
		ExecutionTimeMs:     e.ExecutionTimeMs,
		FirstTokenLatencyMs: e.FirstTokenLatencyMs,
		CreatedAt:           e.CreatedAt,
		CompletedAt:         e.CompletedAt,
	}
	if e.AgentInput != "" {
		if err := json.Unmarshal([]byte(e.AgentInput), &m.AgentInput); err != nil {
			return nil, err
		}
	}
	return m, nil
}
