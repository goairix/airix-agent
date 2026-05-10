package modules

import (
	"github.com/google/wire"

	appService "github.com/dysodeng/app/internal/application/workspace/service"
	domainService "github.com/dysodeng/app/internal/domain/workspace/service"
	wsRepository "github.com/dysodeng/app/internal/infrastructure/persistence/repository/workspace"
	wsHandler "github.com/dysodeng/app/internal/interfaces/http/handler/workspace"
)

// WorkspaceModuleSet 工作空间模块依赖注入聚合
var WorkspaceModuleSet = wire.NewSet(
	// 仓储层
	wsRepository.NewWorkspaceRepository,

	// 领域层
	domainService.NewWorkspaceDomainService,

	// 应用层
	appService.NewWorkspaceApplicationService,

	// 控制器层
	wsHandler.NewWorkspaceHandler,
)
