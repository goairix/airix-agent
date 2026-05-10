package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/application/workspace/dto/command"
	"github.com/dysodeng/app/internal/application/workspace/dto/response"
	wsModel "github.com/dysodeng/app/internal/domain/workspace/model"
	domainService "github.com/dysodeng/app/internal/domain/workspace/service"
	"github.com/dysodeng/app/internal/infrastructure/pkg/logger"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

// Service 工作空间应用服务接口
type Service interface {
	CreateWorkspace(ctx context.Context, cmd *command.CreateWorkspaceCommand) (*response.WorkspaceResponse, error)
	GetWorkspace(ctx context.Context, workspaceID string) (*response.WorkspaceResponse, error)
	ListWorkspaces(ctx context.Context) (*response.WorkspaceListResponse, error)
	DisableWorkspace(ctx context.Context, workspaceID string) error
	EnableWorkspace(ctx context.Context, workspaceID string) error
	AssignAdmin(ctx context.Context, cmd *command.AssignAdminCommand) error
	RevokeAdmin(ctx context.Context, workspaceID, userID string) error
	ListMembers(ctx context.Context, workspaceID string) (*response.MemberListResponse, error)
	GetUserWorkspaces(ctx context.Context, userID string) (*response.WorkspaceListResponse, error)
}

type workspaceApplicationService struct {
	baseTraceSpanName string
	domainService     domainService.Service
}

func NewWorkspaceApplicationService(domainSvc domainService.Service) Service {
	return &workspaceApplicationService{
		baseTraceSpanName: "application.workspace.service.WorkspaceApplicationService",
		domainService:     domainSvc,
	}
}

func (svc *workspaceApplicationService) CreateWorkspace(ctx context.Context, cmd *command.CreateWorkspaceCommand) (*response.WorkspaceResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".CreateWorkspace")
	defer span.End()

	createdBy, err := uuid.Parse(cmd.CreatedBy)
	if err != nil {
		return nil, errors.New("创建人ID格式错误")
	}

	ws, err := svc.domainService.Create(spanCtx, cmd.Name, cmd.Description, createdBy)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toWorkspaceResponse(ws), nil
}

func (svc *workspaceApplicationService) GetWorkspace(ctx context.Context, workspaceID string) (*response.WorkspaceResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetWorkspace")
	defer span.End()

	id, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, errors.New("工作空间ID格式错误")
	}

	ws, err := svc.domainService.GetByID(spanCtx, id)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toWorkspaceResponse(ws), nil
}

func (svc *workspaceApplicationService) ListWorkspaces(ctx context.Context) (*response.WorkspaceListResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListWorkspaces")
	defer span.End()

	list, err := svc.domainService.List(spanCtx)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}

	records := make([]response.WorkspaceResponse, len(list))
	for i, ws := range list {
		wsCopy := ws
		records[i] = *toWorkspaceResponse(&wsCopy)
	}
	return &response.WorkspaceListResponse{Record: records, Total: len(records)}, nil
}

func (svc *workspaceApplicationService) DisableWorkspace(ctx context.Context, workspaceID string) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".DisableWorkspace")
	defer span.End()

	id, err := uuid.Parse(workspaceID)
	if err != nil {
		return errors.New("工作空间ID格式错误")
	}
	if err = svc.domainService.Disable(spanCtx, id); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return err
	}
	return nil
}

func (svc *workspaceApplicationService) EnableWorkspace(ctx context.Context, workspaceID string) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".EnableWorkspace")
	defer span.End()

	id, err := uuid.Parse(workspaceID)
	if err != nil {
		return errors.New("工作空间ID格式错误")
	}
	if err = svc.domainService.Enable(spanCtx, id); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return err
	}
	return nil
}

func (svc *workspaceApplicationService) AssignAdmin(ctx context.Context, cmd *command.AssignAdminCommand) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".AssignAdmin")
	defer span.End()

	workspaceID, err := uuid.Parse(cmd.WorkspaceID)
	if err != nil {
		return errors.New("工作空间ID格式错误")
	}
	userID, err := uuid.Parse(cmd.UserID)
	if err != nil {
		return errors.New("用户ID格式错误")
	}
	if err = svc.domainService.AssignAdmin(spanCtx, workspaceID, userID); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return err
	}
	return nil
}

func (svc *workspaceApplicationService) RevokeAdmin(ctx context.Context, workspaceID, userID string) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".RevokeAdmin")
	defer span.End()

	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		return errors.New("工作空间ID格式错误")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("用户ID格式错误")
	}
	if err = svc.domainService.RevokeAdmin(spanCtx, wsID, uid); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return err
	}
	return nil
}

func (svc *workspaceApplicationService) ListMembers(ctx context.Context, workspaceID string) (*response.MemberListResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListMembers")
	defer span.End()

	id, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, errors.New("工作空间ID格式错误")
	}
	members, err := svc.domainService.ListMembers(spanCtx, id)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	records := make([]response.MemberResponse, len(members))
	for i, m := range members {
		mCopy := m
		records[i] = toMemberResponse(&mCopy)
	}
	return &response.MemberListResponse{Record: records, Total: len(records)}, nil
}

func (svc *workspaceApplicationService) GetUserWorkspaces(ctx context.Context, userID string) (*response.WorkspaceListResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetUserWorkspaces")
	defer span.End()

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("用户ID格式错误")
	}
	members, err := svc.domainService.GetUserWorkspaces(spanCtx, uid)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	records := make([]response.WorkspaceResponse, 0, len(members))
	for _, m := range members {
		records = append(records, response.WorkspaceResponse{ID: m.WorkspaceID.String()})
	}
	return &response.WorkspaceListResponse{Record: records, Total: len(records)}, nil
}

// --- 转换辅助函数 ---

func toWorkspaceResponse(ws *wsModel.Workspace) *response.WorkspaceResponse {
	return &response.WorkspaceResponse{
		ID:          ws.ID.String(),
		Name:        ws.Name,
		Description: ws.Description,
		Status:      ws.Status.Uint8(),
		StatusText:  ws.Status.String(),
		CreatedBy:   ws.CreatedBy.String(),
		CreatedAt:   ws.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toMemberResponse(m *wsModel.Member) response.MemberResponse {
	return response.MemberResponse{
		ID:          m.ID.String(),
		WorkspaceID: m.WorkspaceID.String(),
		UserID:      m.UserID.String(),
		Role:        m.Role.Uint8(),
		RoleText:    m.Role.String(),
		AssignedAt:  m.AssignedAt.Format("2006-01-02 15:04:05"),
	}
}
