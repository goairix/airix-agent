package model

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/dysodeng/app/internal/application/model/dto/command"
	appService "github.com/dysodeng/app/internal/application/model/service"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
	modelRequest "github.com/dysodeng/app/internal/interfaces/http/dto/request/model"
	"github.com/dysodeng/app/internal/interfaces/http/dto/response/api"
	"github.com/dysodeng/app/internal/interfaces/http/validator"
)

// ProviderHandler 模型厂商控制器
type ProviderHandler struct {
	baseTraceSpanName string
	modelService      appService.Service
}

func NewProviderHandler(modelService appService.Service) *ProviderHandler {
	return &ProviderHandler{
		baseTraceSpanName: "interfaces.http.handler.model.ProviderHandler",
		modelService:      modelService,
	}
}

func (h *ProviderHandler) CreateProvider(ctx *gin.Context) {
	spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".CreateProvider")
	defer span.End()

	var req modelRequest.CreateProviderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, validator.TransError(err), api.CodeFail))
		return
	}

	res, err := h.modelService.CreateProvider(spanCtx, &command.CreateProviderCommand{
		Name:         req.Name,
		Protocol:     req.Protocol,
		BaseURL:      req.BaseURL,
		AuthType:     req.AuthType,
		Capabilities: req.Capabilities,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
		return
	}
	ctx.JSON(http.StatusOK, api.Success(spanCtx, res))
}

func (h *ProviderHandler) GetProvider(ctx *gin.Context) {
	spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".GetProvider")
	defer span.End()

	res, err := h.modelService.GetProvider(spanCtx, ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
		return
	}
	ctx.JSON(http.StatusOK, api.Success(spanCtx, res))
}

func (h *ProviderHandler) ListProviders(ctx *gin.Context) {
	spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".ListProviders")
	defer span.End()

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	res, err := h.modelService.ListProviders(spanCtx, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
		return
	}
	ctx.JSON(http.StatusOK, api.Success(spanCtx, res))
}

func (h *ProviderHandler) UpdateProvider(ctx *gin.Context) {
	spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".UpdateProvider")
	defer span.End()

	var req modelRequest.UpdateProviderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, validator.TransError(err), api.CodeFail))
		return
	}

	res, err := h.modelService.UpdateProvider(spanCtx, &command.UpdateProviderCommand{
		ProviderID:   ctx.Param("id"),
		Name:         req.Name,
		Protocol:     req.Protocol,
		BaseURL:      req.BaseURL,
		AuthType:     req.AuthType,
		Capabilities: req.Capabilities,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
		return
	}
	ctx.JSON(http.StatusOK, api.Success(spanCtx, res))
}

func (h *ProviderHandler) DeleteProvider(ctx *gin.Context) {
	spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".DeleteProvider")
	defer span.End()

	if err := h.modelService.DeleteProvider(spanCtx, ctx.Param("id")); err != nil {
		ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
		return
	}
	ctx.JSON(http.StatusOK, api.Success[any](spanCtx, nil))
}
