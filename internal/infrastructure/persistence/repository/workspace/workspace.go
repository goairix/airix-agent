package workspace

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	wsModel "github.com/dysodeng/app/internal/domain/workspace/model"
	"github.com/dysodeng/app/internal/domain/workspace/repository"
	"github.com/dysodeng/app/internal/domain/workspace/valueobject"
	wsEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/workspace"
	"github.com/dysodeng/app/internal/infrastructure/persistence/transactions"
	pkgModel "github.com/dysodeng/app/internal/infrastructure/pkg/model"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

type workspaceRepository struct {
	baseTraceSpanName string
	txManager         transactions.TransactionManager
}

func NewWorkspaceRepository(txManager transactions.TransactionManager) repository.Repository {
	return &workspaceRepository{
		baseTraceSpanName: "infrastructure.persistence.repository.workspace.Repository",
		txManager:         txManager,
	}
}

func (repo *workspaceRepository) Save(ctx context.Context, w *wsModel.Workspace) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Save")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	entity := repo.toEntity(w)

	if w.ID != uuid.Nil {
		var exists wsEntity.Workspace
		tx.Where("id = ?", entity.ID).First(&exists)
		if exists.ID == uuid.Nil {
			if err := tx.Create(entity).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Where("id = ?", entity.ID).Updates(entity).Error; err != nil {
				return err
			}
		}
	} else {
		if err := tx.Create(entity).Error; err != nil {
			return err
		}
		w.ID = entity.ID
		w.CreatedAt = entity.CreatedAt.Time
	}
	return nil
}

func (repo *workspaceRepository) FindByID(ctx context.Context, id uuid.UUID) (*wsModel.Workspace, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindByID")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entity wsEntity.Workspace
	if err := tx.Where("id = ?", id).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return repo.fromEntity(&entity), nil
}

func (repo *workspaceRepository) FindAll(ctx context.Context) ([]wsModel.Workspace, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindAll")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entities []wsEntity.Workspace
	if err := tx.Order("created_at DESC").Find(&entities).Error; err != nil {
		return nil, err
	}
	result := make([]wsModel.Workspace, len(entities))
	for i, e := range entities {
		result[i] = *repo.fromEntity(&e)
	}
	return result, nil
}

func (repo *workspaceRepository) SaveMember(ctx context.Context, m *wsModel.Member) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".SaveMember")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	entity := repo.toMemberEntity(m)
	if err := tx.Create(entity).Error; err != nil {
		return err
	}
	m.ID = entity.ID
	return nil
}

func (repo *workspaceRepository) FindMemberByWorkspaceAndUser(ctx context.Context, workspaceID, userID uuid.UUID) (*wsModel.Member, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindMemberByWorkspaceAndUser")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entity wsEntity.Member
	if err := tx.Where("workspace_id = ? AND user_id = ?", workspaceID, userID).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return repo.fromMemberEntity(&entity), nil
}

func (repo *workspaceRepository) FindMembersByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]wsModel.Member, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindMembersByWorkspace")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entities []wsEntity.Member
	if err := tx.Where("workspace_id = ?", workspaceID).Find(&entities).Error; err != nil {
		return nil, err
	}
	result := make([]wsModel.Member, len(entities))
	for i, e := range entities {
		result[i] = *repo.fromMemberEntity(&e)
	}
	return result, nil
}

func (repo *workspaceRepository) FindMembersByUser(ctx context.Context, userID uuid.UUID) ([]wsModel.Member, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindMembersByUser")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entities []wsEntity.Member
	if err := tx.Where("user_id = ?", userID).Find(&entities).Error; err != nil {
		return nil, err
	}
	result := make([]wsModel.Member, len(entities))
	for i, e := range entities {
		result[i] = *repo.fromMemberEntity(&e)
	}
	return result, nil
}

func (repo *workspaceRepository) DeleteMember(ctx context.Context, workspaceID, userID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".DeleteMember")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	return tx.Where("workspace_id = ? AND user_id = ?", workspaceID, userID).Delete(&wsEntity.Member{}).Error
}

// --- 转换方法 ---

func (repo *workspaceRepository) fromEntity(e *wsEntity.Workspace) *wsModel.Workspace {
	return &wsModel.Workspace{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		Status:      valueobject.Status(e.Status),
		CreatedBy:   e.CreatedBy,
		CreatedAt:   e.CreatedAt.Time,
		UpdatedAt:   e.UpdatedAt.Time,
	}
}

func (repo *workspaceRepository) toEntity(w *wsModel.Workspace) *wsEntity.Workspace {
	return &wsEntity.Workspace{
		DistributedPrimaryKeyID: pkgModel.DistributedPrimaryKeyID{ID: w.ID},
		Name:                    w.Name,
		Description:             w.Description,
		Status:                  w.Status.Uint8(),
		CreatedBy:               w.CreatedBy,
	}
}

func (repo *workspaceRepository) fromMemberEntity(e *wsEntity.Member) *wsModel.Member {
	return &wsModel.Member{
		ID:          e.ID,
		WorkspaceID: e.WorkspaceID,
		UserID:      e.UserID,
		Role:        valueobject.MemberRole(e.Role),
		AssignedAt:  e.AssignedAt.Time,
	}
}

func (repo *workspaceRepository) toMemberEntity(m *wsModel.Member) *wsEntity.Member {
	return &wsEntity.Member{
		DistributedPrimaryKeyID: pkgModel.DistributedPrimaryKeyID{ID: m.ID},
		WorkspaceID:             m.WorkspaceID,
		UserID:                  m.UserID,
		Role:                    m.Role.Uint8(),
		AssignedAt:              pkgModel.JSONTime{Time: m.AssignedAt},
	}
}
