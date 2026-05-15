package model

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/dysodeng/app/internal/application/model/dto/command"
	appService "github.com/dysodeng/app/internal/application/model/service"
	bizCtx "github.com/dysodeng/app/internal/infrastructure/pkg/context"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
	modelRequest "github.com/dysodeng/app/internal/interfaces/http/dto/request/model"
	"github.com/dysodeng/app/internal/interfaces/http/dto/response/api"
	"github.com/dysodeng/app/internal/interfaces/http/validator"
)

// InstanceHandler 模型实例控制器
type InstanceHandler struct {
	baseTraceSpanName string
	modelService      appService.Service
}

func NewInstanceHandler(modelService appService.Service) *InstanceHandler {
	return &InstanceHandler{
		baseTraceSpanName: "interfaces.http.handler.model.InstanceHandler",
		modelService:      modelService,
	}
}

func (h *InstanceHandler) CreateInstance(ctx *gin.Context) {
	spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".CreateInstance")
	defer span.End()

	var req modelRequest.CreateInstanceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, validator.TransError(err), api.CodeFail))
		return
	}

	res, err := h.modelService.CreateInstance(spanCtx, &command.CreateInstanceCommand{
		WorkspaceID:  bizCtx.NewBizContext(spanCtx).Data().WorkspaceID,
		ProviderID:   req.ProviderID,
		ModelName:    req.ModelName,
		Capability:   req.Capability,
		APIKey:       req.APIKey,
		Parameters:   req.Parameters,
		RateLimitRPM: req.RateLimitRPM,
		RateLimitTPM: req.RateLimitTPM,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
		return
	}
	ctx.JSON(http.StatusOK, api.Success(spanCtx, res))
}

func (h *InstanceHandler) GetInstance(ctx *gin.Context) {
	spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".GetInstance")
	defer span.End()

	res, err := h.modelService.GetInstance(spanCtx, ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
		return
	}
	ctx.JSON(http.StatusOK, api.Success(spanCtx, res))
}

func (h *InstanceHandler) ListInstances(ctx *gin.Context) {
	spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".ListInstances")
	defer span.End()

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	res, err := h.modelService.ListInstances(spanCtx, bizCtx.NewBizContext(spanCtx).Data().WorkspaceID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
		return
	}
	ctx.JSON(http.StatusOK, api.Success(spanCtx, res))
}

func (h *InstanceHandler) UpdateInstance(ctx *gin.Context) {
	spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".UpdateInstance")
	defer span.End()

	var req modelRequest.UpdateInstanceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, validator.TransError(err), api.CodeFail))
		return
	}

	res, err := h.modelService.UpdateInstance(spanCtx, &command.UpdateInstanceCommand{
		InstanceID:   ctx.Param("id"),
		ModelName:    req.ModelName,
		Capability:   req.Capability,
		APIKey:       req.APIKey,
		Parameters:   req.Parameters,
		RateLimitRPM: req.RateLimitRPM,
		RateLimitTPM: req.RateLimitTPM,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
		return
	}
	ctx.JSON(http.StatusOK, api.Success(spanCtx, res))
}

func (h *InstanceHandler) DeleteInstance(ctx *gin.Context) {
	spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".DeleteInstance")
	defer span.End()

	if err := h.modelService.DeleteInstance(spanCtx, ctx.Param("id")); err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
		return
	}
	ctx.JSON(http.StatusOK, api.Success[any](spanCtx, nil))
}

func (h *InstanceHandler) EnableInstance(ctx *gin.Context) {
	spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".EnableInstance")
	defer span.End()

	if err := h.modelService.EnableInstance(spanCtx, ctx.Param("id")); err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
		return
	}
	ctx.JSON(http.StatusOK, api.Success[any](spanCtx, nil))
}

func (h *InstanceHandler) DisableInstance(ctx *gin.Context) {
	spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".DisableInstance")
	defer span.End()

	if err := h.modelService.DisableInstance(spanCtx, ctx.Param("id")); err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
		return
	}
	ctx.JSON(http.StatusOK, api.Success[any](spanCtx, nil))
}
