package http

import (
	"github.com/dysodeng/app/internal/interfaces/http/handler/file"
	"github.com/dysodeng/app/internal/interfaces/http/handler/passport"
	"github.com/dysodeng/app/internal/interfaces/http/handler/workspace"
)

// HandlerRegistry 控制器注册表
type HandlerRegistry struct {
	PassportHandler  *passport.Handler
	UploaderHandler  *file.UploaderHandler
	WorkspaceHandler *workspace.Handler
}

func NewHandlerRegistry(
	passportHandler *passport.Handler,
	uploaderHandler *file.UploaderHandler,
	workspaceHandler *workspace.Handler,
) *HandlerRegistry {
	return &HandlerRegistry{
		PassportHandler:  passportHandler,
		UploaderHandler:  uploaderHandler,
		WorkspaceHandler: workspaceHandler,
	}
}
