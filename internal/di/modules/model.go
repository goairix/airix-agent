package modules

import (
	"github.com/google/wire"

	appService "github.com/dysodeng/app/internal/application/model/service"
	domainService "github.com/dysodeng/app/internal/domain/model/service"
	modelAdapter "github.com/dysodeng/app/internal/infrastructure/adapter/model"
	"github.com/dysodeng/app/internal/infrastructure/config"
	modelRepository "github.com/dysodeng/app/internal/infrastructure/persistence/repository/model"
)

// ProvideInstanceRepositoryConfig 从应用配置提取 Instance 仓储加密配置
func ProvideInstanceRepositoryConfig(cfg *config.Config) modelRepository.InstanceRepositoryConfig {
	return modelRepository.InstanceRepositoryConfig{
		EncryptKey: []byte(cfg.Security.Crypto.AESKey),
		EncryptIV:  []byte(cfg.Security.Crypto.AESIV),
	}
}

// ModelModuleSet 模型管理模块依赖注入聚合
var ModelModuleSet = wire.NewSet(
	// 仓储层
	modelRepository.NewProviderRepository,
	ProvideInstanceRepositoryConfig,
	modelRepository.NewInstanceRepository,

	// 领域层
	domainService.NewModelDomainService,

	// 应用层
	appService.NewModelApplicationService,

	// 适配层
	modelAdapter.NewAdapterFactory,
	modelAdapter.NewManager,
)
