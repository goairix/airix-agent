package memory

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	memoryModel "github.com/dysodeng/app/internal/domain/memory/model"
	"github.com/dysodeng/app/internal/domain/memory/repository"
	"github.com/dysodeng/app/internal/domain/memory/valueobject"
	memoryEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/memory"
	"github.com/dysodeng/app/internal/infrastructure/persistence/transactions"
	pkgModel "github.com/dysodeng/app/internal/infrastructure/pkg/model"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

type memoryRepository struct {
	baseTraceSpanName string
	txManager         transactions.TransactionManager
}

func NewMemoryRepository(txManager transactions.TransactionManager) repository.MemoryRepository {
	return &memoryRepository{
		baseTraceSpanName: "infrastructure.persistence.repository.memory.MemoryRepository",
		txManager:         txManager,
	}
}

func (repo *memoryRepository) Save(ctx context.Context, m *memoryModel.Memory) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Save")
	defer span.End()
	entity, err := repo.toEntity(m)
	if err != nil {
		return err
	}
	if err = repo.txManager.GetTx(spanCtx).Create(entity).Error; err != nil {
		return err
	}
	m.ID = entity.ID
	return nil
}

func (repo *memoryRepository) ListByUserAndDate(ctx context.Context, workspaceID, userID uuid.UUID, date time.Time) ([]memoryModel.Memory, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".ListByUserAndDate")
	defer span.End()
	var entities []memoryEntity.Memory
	if err := repo.txManager.GetTx(spanCtx).
		Where("workspace_id = ? AND user_id = ? AND date = ?", workspaceID, userID, date.Format("2006-01-02")).
		Find(&entities).Error; err != nil {
		return nil, err
	}
	return repo.fromEntities(entities)
}

func (repo *memoryRepository) ListByAgentUser(ctx context.Context, workspaceID, agentID, userID uuid.UUID) ([]memoryModel.Memory, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".ListByAgentUser")
	defer span.End()
	var entities []memoryEntity.Memory
	if err := repo.txManager.GetTx(spanCtx).
		Where("workspace_id = ? AND agent_id = ? AND user_id = ? AND memory_type = ?", workspaceID, agentID, userID, 1).
		Find(&entities).Error; err != nil {
		return nil, err
	}
	return repo.fromEntities(entities)
}

func (repo *memoryRepository) DeleteByAgentUser(ctx context.Context, workspaceID, agentID, userID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".DeleteByAgentUser")
	defer span.End()
	return repo.txManager.GetTx(spanCtx).
		Where("workspace_id = ? AND agent_id = ? AND user_id = ? AND memory_type = ?", workspaceID, agentID, userID, 1).
		Delete(&memoryEntity.Memory{}).Error
}

func (repo *memoryRepository) toEntity(m *memoryModel.Memory) (*memoryEntity.Memory, error) {
	tagsJSON, err := json.Marshal(m.Tags)
	if err != nil {
		return nil, err
	}
	return &memoryEntity.Memory{
		DistributedPrimaryKeyID: pkgModel.DistributedPrimaryKeyID{ID: m.ID},
		WorkspaceID:             m.WorkspaceID,
		UserID:                  m.UserID,
		MemoryType:              m.MemoryType.Uint8(),
		AgentID:                 m.AgentID,
		Content:                 m.Content,
		Tags:                    string(tagsJSON),
		Importance:              m.Importance,
		Date:                    m.Date,
	}, nil
}

func (repo *memoryRepository) fromEntities(entities []memoryEntity.Memory) ([]memoryModel.Memory, error) {
	memories := make([]memoryModel.Memory, 0, len(entities))
	for _, e := range entities {
		var tags []string
		if e.Tags != "" {
			_ = json.Unmarshal([]byte(e.Tags), &tags)
		}
		memories = append(memories, memoryModel.Memory{
			ID:          e.ID,
			WorkspaceID: e.WorkspaceID,
			UserID:      e.UserID,
			MemoryType:  valueobject.MemoryType(e.MemoryType),
			AgentID:     e.AgentID,
			Content:     e.Content,
			Tags:        tags,
			Importance:  e.Importance,
			Date:        e.Date,
			CreatedAt:   e.CreatedAt.Time,
		})
	}
	return memories, nil
}
