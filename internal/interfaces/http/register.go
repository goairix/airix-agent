package http

import (
	"github.com/dysodeng/app/internal/interfaces/http/handler/file"
	modelHandler "github.com/dysodeng/app/internal/interfaces/http/handler/model"
	"github.com/dysodeng/app/internal/interfaces/http/handler/passport"
	"github.com/dysodeng/app/internal/interfaces/http/handler/workspace"
)

// HandlerRegistry 控制器注册表
type HandlerRegistry struct {
	PassportHandler  *passport.Handler
	UploaderHandler  *file.UploaderHandler
	WorkspaceHandler *workspace.Handler
	ProviderHandler  *modelHandler.ProviderHandler
	InstanceHandler  *modelHandler.InstanceHandler
}

func NewHandlerRegistry(
	passportHandler *passport.Handler,
	uploaderHandler *file.UploaderHandler,
	workspaceHandler *workspace.Handler,
	providerHandler *modelHandler.ProviderHandler,
	instanceHandler *modelHandler.InstanceHandler,
) *HandlerRegistry {
	return &HandlerRegistry{
		PassportHandler:  passportHandler,
		UploaderHandler:  uploaderHandler,
		WorkspaceHandler: workspaceHandler,
		ProviderHandler:  providerHandler,
		InstanceHandler:  instanceHandler,
	}
}
