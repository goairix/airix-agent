package service

import (
	"context"

	"github.com/google/uuid"

	wsErrors "github.com/dysodeng/app/internal/domain/workspace/errors"
	"github.com/dysodeng/app/internal/domain/workspace/model"
	"github.com/dysodeng/app/internal/domain/workspace/repository"
	"github.com/dysodeng/app/internal/domain/workspace/valueobject"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

// Service 工作空间领域服务接口
type Service interface {
	Create(ctx context.Context, name, description string, createdBy uuid.UUID) (*model.Workspace, error)
	GetByID(ctx context.Context, workspaceID uuid.UUID) (*model.Workspace, error)
	List(ctx context.Context) ([]model.Workspace, error)
	Disable(ctx context.Context, workspaceID uuid.UUID) error
	Enable(ctx context.Context, workspaceID uuid.UUID) error
	AssignAdmin(ctx context.Context, workspaceID, userID uuid.UUID) error
	RevokeAdmin(ctx context.Context, workspaceID, userID uuid.UUID) error
	ListMembers(ctx context.Context, workspaceID uuid.UUID) ([]model.Member, error)
	GetUserWorkspaces(ctx context.Context, userID uuid.UUID) ([]model.Member, error)
}

type workspaceDomainService struct {
	baseTraceSpanName string
	repository        repository.Repository
}

func NewWorkspaceDomainService(repo repository.Repository) Service {
	return &workspaceDomainService{
		baseTraceSpanName: "domain.workspace.service.WorkspaceDomainService",
		repository:        repo,
	}
}

func (svc *workspaceDomainService) Create(ctx context.Context, name, description string, createdBy uuid.UUID) (*model.Workspace, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".Create")
	defer span.End()

	ws, err := model.NewWorkspace(name, description, createdBy)
	if err != nil {
		return nil, err
	}
	if err = svc.repository.Save(spanCtx, ws); err != nil {
		return nil, wsErrors.ErrWorkspaceSaveFailed.WrapNew(err)
	}
	return ws, nil
}

func (svc *workspaceDomainService) GetByID(ctx context.Context, workspaceID uuid.UUID) (*model.Workspace, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetByID")
	defer span.End()

	ws, err := svc.repository.FindByID(spanCtx, workspaceID)
	if err != nil {
		return nil, wsErrors.ErrWorkspaceQueryFailed.WrapNew(err)
	}
	if ws == nil || ws.ID == uuid.Nil {
		return nil, wsErrors.ErrWorkspaceNotFound
	}
	return ws, nil
}

func (svc *workspaceDomainService) List(ctx context.Context) ([]model.Workspace, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".List")
	defer span.End()

	list, err := svc.repository.FindAll(spanCtx)
	if err != nil {
		return nil, wsErrors.ErrWorkspaceQueryFailed.WrapNew(err)
	}
	return list, nil
}

func (svc *workspaceDomainService) Disable(ctx context.Context, workspaceID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".Disable")
	defer span.End()

	ws, err := svc.GetByID(spanCtx, workspaceID)
	if err != nil {
		return err
	}
	ws.Disable()
	if err = svc.repository.Save(spanCtx, ws); err != nil {
		return wsErrors.ErrWorkspaceSaveFailed.WrapNew(err)
	}
	return nil
}

func (svc *workspaceDomainService) Enable(ctx context.Context, workspaceID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".Enable")
	defer span.End()

	ws, err := svc.GetByID(spanCtx, workspaceID)
	if err != nil {
		return err
	}
	ws.Enable()
	if err = svc.repository.Save(spanCtx, ws); err != nil {
		return wsErrors.ErrWorkspaceSaveFailed.WrapNew(err)
	}
	return nil
}

func (svc *workspaceDomainService) AssignAdmin(ctx context.Context, workspaceID, userID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".AssignAdmin")
	defer span.End()

	if _, err := svc.GetByID(spanCtx, workspaceID); err != nil {
		return err
	}

	existing, err := svc.repository.FindMemberByWorkspaceAndUser(spanCtx, workspaceID, userID)
	if err != nil {
		return wsErrors.ErrWorkspaceMemberQueryFailed.WrapNew(err)
	}
	if existing != nil && existing.ID != uuid.Nil {
		return wsErrors.ErrWorkspaceMemberExists
	}

	member, err := model.NewMember(workspaceID, userID, valueobject.RoleAdmin)
	if err != nil {
		return err
	}
	if err = svc.repository.SaveMember(spanCtx, member); err != nil {
		return wsErrors.ErrWorkspaceMemberSaveFailed.WrapNew(err)
	}
	return nil
}

func (svc *workspaceDomainService) RevokeAdmin(ctx context.Context, workspaceID, userID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".RevokeAdmin")
	defer span.End()

	existing, err := svc.repository.FindMemberByWorkspaceAndUser(spanCtx, workspaceID, userID)
	if err != nil {
		return wsErrors.ErrWorkspaceMemberQueryFailed.WrapNew(err)
	}
	if existing == nil || existing.ID == uuid.Nil {
		return wsErrors.ErrWorkspaceMemberNotFound
	}

	if err = svc.repository.DeleteMember(spanCtx, workspaceID, userID); err != nil {
		return wsErrors.ErrWorkspaceMemberDeleteFailed.WrapNew(err)
	}
	return nil
}

func (svc *workspaceDomainService) ListMembers(ctx context.Context, workspaceID uuid.UUID) ([]model.Member, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListMembers")
	defer span.End()

	members, err := svc.repository.FindMembersByWorkspace(spanCtx, workspaceID)
	if err != nil {
		return nil, wsErrors.ErrWorkspaceMemberQueryFailed.WrapNew(err)
	}
	return members, nil
}

func (svc *workspaceDomainService) GetUserWorkspaces(ctx context.Context, userID uuid.UUID) ([]model.Member, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetUserWorkspaces")
	defer span.End()

	members, err := svc.repository.FindMembersByUser(spanCtx, userID)
	if err != nil {
		return nil, wsErrors.ErrWorkspaceMemberQueryFailed.WrapNew(err)
	}
	return members, nil
}
