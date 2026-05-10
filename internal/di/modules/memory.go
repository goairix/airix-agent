package modules

import (
	"github.com/google/wire"

	appService "github.com/dysodeng/app/internal/application/memory/service"
	memoryRepository "github.com/dysodeng/app/internal/infrastructure/persistence/repository/memory"
)

// MemoryModuleSet Memory 模块依赖注入聚合
var MemoryModuleSet = wire.NewSet(
	memoryRepository.NewMemoryRepository,
	appService.NewMemoryApplicationService,
)
