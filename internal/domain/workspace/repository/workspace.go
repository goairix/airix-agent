package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/domain/workspace/model"
)

// Repository 工作空间仓储接口
type Repository interface {
	Save(ctx context.Context, workspace *model.Workspace) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Workspace, error)
	FindAll(ctx context.Context) ([]model.Workspace, error)
	SaveMember(ctx context.Context, member *model.Member) error
	FindMemberByWorkspaceAndUser(ctx context.Context, workspaceID, userID uuid.UUID) (*model.Member, error)
	FindMembersByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]model.Member, error)
	FindMembersByUser(ctx context.Context, userID uuid.UUID) ([]model.Member, error)
	DeleteMember(ctx context.Context, workspaceID, userID uuid.UUID) error
}
