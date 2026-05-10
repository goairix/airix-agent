package modules

import (
	"github.com/google/wire"

	appService "github.com/dysodeng/app/internal/application/session/service"
	sessionRepository "github.com/dysodeng/app/internal/infrastructure/persistence/repository/session"
)

// SessionModuleSet Session 模块依赖注入聚合
// 注意：SlidingWindowAssembler 不在此注入，因其 windowSize 来自 Agent 运行时配置，
// 由应用服务在调用时动态构造：appService.NewSlidingWindowAssembler(cfg.WindowSize, msgRepo, itemRepo)
var SessionModuleSet = wire.NewSet(
	sessionRepository.NewSessionRepository,
	sessionRepository.NewMessageRepository,
	sessionRepository.NewMessageItemRepository,
	appService.NewSessionApplicationService,
)
