package agent

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	agentModel "github.com/dysodeng/app/internal/domain/agent/model"
	"github.com/dysodeng/app/internal/domain/agent/repository"
	"github.com/dysodeng/app/internal/domain/agent/valueobject"
	agentEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/agent"
	"github.com/dysodeng/app/internal/infrastructure/persistence/transactions"
	pkgModel "github.com/dysodeng/app/internal/infrastructure/pkg/model"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

type agentReleaseRepository struct {
	baseTraceSpanName string
	txManager         transactions.TransactionManager
}

func NewAgentReleaseRepository(txManager transactions.TransactionManager) repository.ReleaseRepository {
	return &agentReleaseRepository{
		baseTraceSpanName: "infrastructure.persistence.repository.agent.ReleaseRepository",
		txManager:         txManager,
	}
}

func (repo *agentReleaseRepository) Save(ctx context.Context, r *agentModel.AgentRelease) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Save")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	entity, err := repo.toEntity(r)
	if err != nil {
		return err
	}
	if err = tx.Create(entity).Error; err != nil {
		return err
	}
	r.ReleasedAt = entity.ReleasedAt.Time
	return nil
}

func (repo *agentReleaseRepository) FindByID(ctx context.Context, releaseID string) (*agentModel.AgentRelease, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindByID")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entity agentEntity.AgentRelease
	if err := tx.Where("release_id = ?", releaseID).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return repo.fromEntity(&entity)
}

func (repo *agentReleaseRepository) FindActive(ctx context.Context, agentID uuid.UUID) (*agentModel.AgentRelease, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindActive")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entity agentEntity.AgentRelease
	if err := tx.Where("agent_id = ? AND status = ?", agentID, valueobject.ReleaseStatusActive.Uint8()).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return repo.fromEntity(&entity)
}

func (repo *agentReleaseRepository) DeactivateAll(ctx context.Context, agentID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".DeactivateAll")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	return tx.Model(&agentEntity.AgentRelease{}).
		Where("agent_id = ? AND status = ?", agentID, valueobject.ReleaseStatusActive.Uint8()).
		Update("status", valueobject.ReleaseStatusInactive.Uint8()).Error
}

func (repo *agentReleaseRepository) ListByAgent(ctx context.Context, agentID uuid.UUID, pagination repository.Pagination) ([]agentModel.AgentRelease, int64, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".ListByAgent")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entities []agentEntity.AgentRelease
	var total int64

	query := tx.Model(&agentEntity.AgentRelease{}).Where("agent_id = ?", agentID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	if err := query.Order("released_at DESC").Offset(offset).Limit(pagination.PageSize).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	result := make([]agentModel.AgentRelease, 0, len(entities))
	for _, e := range entities {
		r, err := repo.fromEntity(&e)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, *r)
	}
	return result, total, nil
}

// --- 转换方法 ---

func (repo *agentReleaseRepository) fromEntity(e *agentEntity.AgentRelease) (*agentModel.AgentRelease, error) {
	var snapshot agentModel.AgentSnapshot
	if e.Snapshot != "" {
		if err := json.Unmarshal([]byte(e.Snapshot), &snapshot); err != nil {
			return nil, err
		}
	}
	return &agentModel.AgentRelease{
		ReleaseID:   e.ReleaseID,
		AgentID:     e.AgentID,
		WorkspaceID: e.WorkspaceID,
		ReleasedAt:  e.ReleasedAt.Time,
		ReleasedBy:  e.ReleasedBy,
		ChangeLog:   e.ChangeLog,
		Status:      valueobject.ReleaseStatus(e.Status),
		Snapshot:    snapshot,
	}, nil
}

func (repo *agentReleaseRepository) toEntity(r *agentModel.AgentRelease) (*agentEntity.AgentRelease, error) {
	snapshotJSON, err := json.Marshal(r.Snapshot)
	if err != nil {
		return nil, err
	}
	releasedAt := r.ReleasedAt
	if releasedAt.IsZero() {
		releasedAt = time.Now()
	}
	return &agentEntity.AgentRelease{
		DistributedPrimaryKeyID: pkgModel.DistributedPrimaryKeyID{},
		ReleaseID:               r.ReleaseID,
		AgentID:                 r.AgentID,
		WorkspaceID:             r.WorkspaceID,
		ReleasedAt:              pkgModel.JSONTime{Time: releasedAt},
		ReleasedBy:              r.ReleasedBy,
		ChangeLog:               r.ChangeLog,
		Status:                  r.Status.Uint8(),
		Snapshot:                string(snapshotJSON),
	}, nil
}
