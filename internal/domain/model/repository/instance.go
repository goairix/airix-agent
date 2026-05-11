// internal/domain/model/repository/instance.go
package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/domain/model/model"
	"github.com/dysodeng/app/internal/domain/model/valueobject"
)

// InstanceRepository 模型实例仓储接口
type InstanceRepository interface {
	Save(ctx context.Context, instance *model.Instance) error
	FindByID(ctx context.Context, instanceID uuid.UUID) (*model.Instance, error)
	FindByWorkspace(ctx context.Context, workspaceID uuid.UUID, pagination Pagination) ([]model.Instance, int64, error)
	FindByWorkspaceAndCapability(ctx context.Context, workspaceID uuid.UUID, capability valueobject.Capability, pagination Pagination) ([]model.Instance, int64, error)
	ExistsByProviderID(ctx context.Context, providerID uuid.UUID) (bool, error)
	Delete(ctx context.Context, instanceID uuid.UUID) error
}
