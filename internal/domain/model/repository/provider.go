// internal/domain/model/repository/provider.go
package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/domain/model/model"
)

// Pagination 分页参数
type Pagination struct {
	Page     int
	PageSize int
}

// ProviderRepository 模型厂商仓储接口
type ProviderRepository interface {
	Save(ctx context.Context, provider *model.Provider) error
	FindByID(ctx context.Context, providerID uuid.UUID) (*model.Provider, error)
	FindAll(ctx context.Context, pagination Pagination) ([]model.Provider, int64, error)
	Delete(ctx context.Context, providerID uuid.UUID) error
}
