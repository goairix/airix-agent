package workspace

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dysodeng/app/internal/application/workspace/dto/command"
	appService "github.com/dysodeng/app/internal/application/workspace/service"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
	wsRequest "github.com/dysodeng/app/internal/interfaces/http/dto/request/workspace"
	"github.com/dysodeng/app/internal/interfaces/http/dto/response/api"
	"github.com/dysodeng/app/internal/interfaces/http/validator"
)

// Handler 工作空间控制器
type Handler struct {
	baseTraceSpanName string
	workspaceService  appService.Service
}

func NewWorkspaceHandler(workspaceService appService.Service) *Handler {
	return &Handler{
		baseTraceSpanName: "interfaces.http.handler.workspace.Handler",
		workspaceService:  workspaceService,
	}
}

// CreateWorkspace 创建工作空间
func (h *Handler) CreateWorkspace(ctx *gin.Context) {
	spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".CreateWorkspace")
	defer span.End()

	var req wsRequest.CreateWorkspaceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, validator.TransError(err), api.CodeFail))
		return
	}

	res, err := h.workspaceService.CreateWorkspace(spanCtx, &command.CreateWorkspaceCommand{
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   req.CreatedBy,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
		return
	}
	ctx.JSON(http.StatusOK, api.Success(spanCtx, res))
}

// GetWorkspace 获取工作空间详情
func (h *Handler) GetWorkspace(ctx *gin.Context) {
	spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".GetWorkspace")
	defer span.End()

	workspaceID := ctx.Param("id")
	res, err := h.workspaceService.GetWorkspace(spanCtx, workspaceID)
	if err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
		return
	}
	ctx.JSON(http.StatusOK, api.Success(spanCtx, res))
}

// ListWorkspaces 工作空间列表
func (h *Handler) ListWorkspaces(ctx *gin.Context) {
	spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".ListWorkspaces")
	defer span.End()

	res, err := h.workspaceService.ListWorkspaces(spanCtx)
	if err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
		return
	}
	ctx.JSON(http.StatusOK, api.Success(spanCtx, res))
}

// DisableWorkspace 禁用工作空间
func (h *Handler) DisableWorkspace(ctx *gin.Context) {
	spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".DisableWorkspace")
	defer span.End()

	if err := h.workspaceService.DisableWorkspace(spanCtx, ctx.Param("id")); err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
		return
	}
	ctx.JSON(http.StatusOK, api.Success[any](spanCtx, nil))
}

// EnableWorkspace 启用工作空间
func (h *Handler) EnableWorkspace(ctx *gin.Context) {
	spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".EnableWorkspace")
	defer span.End()

	if err := h.workspaceService.EnableWorkspace(spanCtx, ctx.Param("id")); err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
		return
	}
	ctx.JSON(http.StatusOK, api.Success[any](spanCtx, nil))
}

// AssignAdmin 分配管理员
func (h *Handler) AssignAdmin(ctx *gin.Context) {
	spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".AssignAdmin")
	defer span.End()

	var req wsRequest.AssignAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, validator.TransError(err), api.CodeFail))
		return
	}

	if err := h.workspaceService.AssignAdmin(spanCtx, &command.AssignAdminCommand{
		WorkspaceID: ctx.Param("id"),
		UserID:      req.UserID,
	}); err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
		return
	}
	ctx.JSON(http.StatusOK, api.Success[any](spanCtx, nil))
}

// RevokeAdmin 撤销管理员
func (h *Handler) RevokeAdmin(ctx *gin.Context) {
	spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".RevokeAdmin")
	defer span.End()

	if err := h.workspaceService.RevokeAdmin(spanCtx, ctx.Param("id"), ctx.Param("userId")); err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
		return
	}
	ctx.JSON(http.StatusOK, api.Success[any](spanCtx, nil))
}

// ListMembers 工作空间成员列表
func (h *Handler) ListMembers(ctx *gin.Context) {
	spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".ListMembers")
	defer span.End()

	res, err := h.workspaceService.ListMembers(spanCtx, ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
		return
	}
	ctx.JSON(http.StatusOK, api.Success(spanCtx, res))
}
