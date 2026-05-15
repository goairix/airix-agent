package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	bizCtx "github.com/dysodeng/app/internal/infrastructure/pkg/context"
	"github.com/dysodeng/app/internal/interfaces/http/dto/response/api"
)

func WorkspaceRequired(ctx *gin.Context) {
	workspaceId := ctx.Request.Header.Get(bizCtx.WorkspaceIdKey)
	if workspaceId == "" {
		ctx.AbortWithStatusJSON(http.StatusOK, api.Fail(ctx, "缺少工作空间ID", api.CodeUnauthorized))
		return
	}
	wsId, err := uuid.Parse(workspaceId)
	if err != nil || wsId == uuid.Nil {
		ctx.AbortWithStatusJSON(http.StatusOK, api.Fail(ctx, "工作空间ID错误", api.CodeUnauthorized))
		return
	}

	c := context.WithValue(ctx.Request.Context(), bizCtx.WorkspaceIdKey, workspaceId)
	ctx.Request = ctx.Request.WithContext(c)

	ctx.Next()
}
