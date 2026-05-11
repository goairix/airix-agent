package model

import (
	"github.com/dysodeng/app/internal/infrastructure/pkg/model"
)

// Provider 模型厂商数据实体
type Provider struct {
	model.DistributedPrimaryKeyID
	Name                  string `gorm:"type:varchar(100);not null;default:'';comment:厂商名称" json:"name"`
	Protocol              uint8  `gorm:"type:tinyint(3);not null;default:1;comment:协议类型 1-openai-compatible 2-anthropic 3-google 4-custom" json:"protocol"`
	BaseURL               string `gorm:"type:varchar(500);not null;default:'';comment:基础URL" json:"base_url"`
	AuthType              uint8  `gorm:"type:tinyint(3);not null;default:0;comment:认证类型 0-none 1-api-key 2-oauth" json:"auth_type"`
	SupportedCapabilities string `gorm:"type:json;not null;comment:支持的能力列表 JSON" json:"supported_capabilities"`
	model.Time
	model.SoftDelete
}

func (Provider) TableName() string {
	return "model_providers"
}
