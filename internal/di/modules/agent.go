package modules

import (
	"github.com/google/wire"

	memoryAppService "github.com/dysodeng/app/internal/application/agent/memory/service"
	appService "github.com/dysodeng/app/internal/application/agent/service"
	sessionAppService "github.com/dysodeng/app/internal/application/agent/session/service"
	domainService "github.com/dysodeng/app/internal/domain/agent/service"
	agentRepository "github.com/dysodeng/app/internal/infrastructure/persistence/repository/agent"
	memoryRepository "github.com/dysodeng/app/internal/infrastructure/persistence/repository/agent/memory"
	sessionRepository "github.com/dysodeng/app/internal/infrastructure/persistence/repository/agent/session"
)

// AgentModuleSet Agent 模块依赖注入聚合（含 Session/Memory 子域）
// 注意：SlidingWindowAssembler 不在此注入，因其 windowSize 来自 Agent 运行时配置，
// 由应用服务在调用时动态构造：sessionAppService.NewSlidingWindowAssembler(cfg.WindowSize, msgRepo, itemRepo)
var AgentModuleSet = wire.NewSet(
	// Agent 仓储层
	agentRepository.NewAgentRepository,
	agentRepository.NewAgentReleaseRepository,

	// Agent 领域层
	domainService.NewAgentDomainService,

	// Agent 应用层
	appService.NewAgentApplicationService,

	// Session 子域仓储层
	sessionRepository.NewSessionRepository,
	sessionRepository.NewMessageRepository,
	sessionRepository.NewMessageItemRepository,

	// Session 子域应用层
	sessionAppService.NewSessionApplicationService,
	sessionAppService.NewContextAssemblerFactory,

	// Memory 子域仓储层
	memoryRepository.NewMemoryRepository,

	// Memory 子域应用层
	memoryAppService.NewMemoryApplicationService,
)
