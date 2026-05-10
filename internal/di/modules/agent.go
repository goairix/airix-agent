package modules

import (
	"github.com/google/wire"

	appService "github.com/dysodeng/app/internal/application/agent/service"
	domainService "github.com/dysodeng/app/internal/domain/agent/service"
	agentRepository "github.com/dysodeng/app/internal/infrastructure/persistence/repository/agent"
)

// AgentModuleSet Agent 模块依赖注入聚合
var AgentModuleSet = wire.NewSet(
	// 仓储层
	agentRepository.NewAgentRepository,
	agentRepository.NewAgentReleaseRepository,

	// 领域层
	domainService.NewAgentDomainService,

	// 应用层
	appService.NewAgentApplicationService,
)
