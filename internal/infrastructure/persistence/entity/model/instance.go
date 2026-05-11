package model

import (
	"github.com/dysodeng/app/internal/infrastructure/pkg/model"
	"github.com/google/uuid"
)

// Instance 模型实例数据实体
type Instance struct {
	model.DistributedPrimaryKeyID
	WorkspaceID  uuid.UUID `gorm:"type:uuid;not null;index:model_instance_workspace_idx;comment:工作空间ID" json:"workspace_id"`
	ProviderID   uuid.UUID `gorm:"type:uuid;not null;index:model_instance_provider_idx;comment:模型厂商ID" json:"provider_id"`
	ModelName    string    `gorm:"type:varchar(100);not null;default:'';comment:模型名称" json:"model_name"`
	Capability   uint8     `gorm:"type:tinyint(3);not null;default:1;comment:模型能力 1-chat 2-embedding 3-rerank 4-tts 5-stt" json:"capability"`
	APIKey       string    `gorm:"type:text;not null;comment:API Key（AES加密存储）" json:"api_key"`
	Parameters   string    `gorm:"type:json;not null;comment:默认参数 JSON" json:"parameters"`
	RateLimitRPM int       `gorm:"type:int;not null;default:0;comment:每分钟请求数限制" json:"rate_limit_rpm"`
	RateLimitTPM int       `gorm:"type:int;not null;default:0;comment:每分钟Token数限制" json:"rate_limit_tpm"`
	Status       uint8     `gorm:"type:tinyint(3);not null;default:1;comment:状态 1-active 2-disabled" json:"status"`
	model.Time
	model.SoftDelete
}

func (Instance) TableName() string {
	return "model_instances"
}
