// internal/domain/model/model/instance.go
package model

import (
	"time"

	"github.com/google/uuid"

	modelErrors "github.com/dysodeng/app/internal/domain/model/errors"
	"github.com/dysodeng/app/internal/domain/model/valueobject"
)

// RateLimit 速率限制配置
type RateLimit struct {
	RPM int // 每分钟请求数
	TPM int // 每分钟 Token 数
}

// Instance 模型实例聚合根（工作空间级）
type Instance struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	ProviderID  uuid.UUID
	ModelName   string
	Capability  valueobject.Capability
	APIKey      string         // 明文，持久化时加密
	Parameters  map[string]any // 默认参数（temperature, max_tokens 等）
	RateLimit   RateLimit
	Status      valueobject.InstanceStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewInstance(workspaceID, providerID uuid.UUID, modelName string, capability valueobject.Capability) (*Instance, error) {
	id, _ := uuid.NewV7()
	inst := &Instance{
		ID:          id,
		WorkspaceID: workspaceID,
		ProviderID:  providerID,
		ModelName:   modelName,
		Capability:  capability,
		Status:      valueobject.InstanceStatusActive,
	}
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst, nil
}

func (i *Instance) Validate() error {
	if i.WorkspaceID == uuid.Nil {
		return modelErrors.ErrInstanceWorkspaceEmpty
	}
	if i.ProviderID == uuid.Nil {
		return modelErrors.ErrInstanceProviderEmpty
	}
	if i.ModelName == "" {
		return modelErrors.ErrInstanceModelNameEmpty
	}
	if err := i.Capability.Validate(); err != nil {
		return modelErrors.ErrInstanceCapabilityInvalid
	}
	return nil
}

func (i *Instance) Disable() {
	i.Status = valueobject.InstanceStatusDisabled
}

func (i *Instance) Enable() {
	i.Status = valueobject.InstanceStatusActive
}

// SetAPIKey 设置 API Key（明文，持久化时由仓储层加密）
func (i *Instance) SetAPIKey(apiKey string) {
	i.APIKey = apiKey
}

// IsActive 是否处于启用状态
func (i *Instance) IsActive() bool {
	return i.Status == valueobject.InstanceStatusActive
}
