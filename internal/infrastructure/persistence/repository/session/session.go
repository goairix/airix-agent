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

type sessionRepository struct {
	baseTraceSpanName string
	txManager         transactions.TransactionManager
}

func NewSessionRepository(txManager transactions.TransactionManager) repository.SessionRepository {
	return &sessionRepository{
		baseTraceSpanName: "infrastructure.persistence.repository.session.SessionRepository",
		txManager:         txManager,
	}
}

func (repo *sessionRepository) Save(ctx context.Context, s *sessionModel.Session) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Save")
	defer span.End()
	tx := repo.txManager.GetTx(spanCtx)
	entity, err := repo.toEntity(s)
	if err != nil {
		return err
	}
	if s.ID != uuid.Nil {
		var exists sessionEntity.Session
		tx.Where("id = ?", entity.ID).First(&exists)
		if exists.ID == uuid.Nil {
			return tx.Create(entity).Error
		}
		return tx.Where("id = ?", entity.ID).Updates(entity).Error
	}
	if err = tx.Create(entity).Error; err != nil {
		return err
	}
	s.ID = entity.ID
	return nil
}

func (repo *sessionRepository) FindByID(ctx context.Context, sessionID uuid.UUID) (*sessionModel.Session, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindByID")
	defer span.End()
	tx := repo.txManager.GetTx(spanCtx)
	var entity sessionEntity.Session
	if err := tx.Where("id = ?", sessionID).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return repo.fromEntity(&entity)
}

func (repo *sessionRepository) ListByAgent(ctx context.Context, agentID uuid.UUID, pagination repository.Pagination) ([]sessionModel.Session, int64, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".ListByAgent")
	defer span.End()
	tx := repo.txManager.GetTx(spanCtx)
	var entities []sessionEntity.Session
	var total int64
	db := tx.Model(&sessionEntity.Session{}).Where("agent_id = ?", agentID)
	db.Count(&total)
	offset := (pagination.Page - 1) * pagination.PageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pagination.PageSize).Find(&entities).Error; err != nil {
		return nil, 0, err
	}
	sessions := make([]sessionModel.Session, 0, len(entities))
	for _, e := range entities {
		s, err := repo.fromEntity(&e)
		if err != nil {
			return nil, 0, err
		}
		sessions = append(sessions, *s)
	}
	return sessions, total, nil
}

func (repo *sessionRepository) toEntity(s *sessionModel.Session) (*sessionEntity.Session, error) {
	var interruptJSON string
	if s.InterruptState != nil {
		b, err := json.Marshal(s.InterruptState)
		if err != nil {
			return nil, err
		}
		interruptJSON = string(b)
	}
	return &sessionEntity.Session{
		DistributedPrimaryKeyID: pkgModel.DistributedPrimaryKeyID{ID: s.ID},
		WorkspaceID:             s.WorkspaceID,
		AgentID:                 s.AgentID,
		ReleaseID:               s.ReleaseID,
		UserID:                  s.UserID,
		Title:                   s.Title,
		Status:                  s.Status.Uint8(),
		InputTokens:             s.TotalTokenUsage.InputTokens,
		OutputTokens:            s.TotalTokenUsage.OutputTokens,
		CachedTokens:            s.TotalTokenUsage.CachedTokens,
		InterruptState:          interruptJSON,
	}, nil
}

func (repo *sessionRepository) fromEntity(e *sessionEntity.Session) (*sessionModel.Session, error) {
	s := &sessionModel.Session{
		ID:          e.ID,
		WorkspaceID: e.WorkspaceID,
		AgentID:     e.AgentID,
		ReleaseID:   e.ReleaseID,
		UserID:      e.UserID,
		Title:       e.Title,
		Status:      valueobject.SessionStatus(e.Status),
		TotalTokenUsage: sessionModel.TokenUsage{
			InputTokens:  e.InputTokens,
			OutputTokens: e.OutputTokens,
			CachedTokens: e.CachedTokens,
		},
		CreatedAt: e.CreatedAt.Time,
		UpdatedAt: e.UpdatedAt.Time,
	}
	if e.InterruptState != "" {
		var state sessionModel.InterruptState
		if err := json.Unmarshal([]byte(e.InterruptState), &state); err != nil {
			return nil, err
		}
		s.InterruptState = &state
	}
	return s, nil
}
