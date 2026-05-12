package model

// CreateProviderRequest 创建模型厂商请求
type CreateProviderRequest struct {
	Name         string  `json:"name" binding:"required,max=100" msg:"厂商名称不能为空"`
	Protocol     uint8   `json:"protocol" binding:"required,min=1,max=4" msg:"协议类型无效"`
	BaseURL      string  `json:"base_url"`
	AuthType     uint8   `json:"auth_type" binding:"max=2"`
	Capabilities []uint8 `json:"capabilities"`
}

// UpdateProviderRequest 更新模型厂商请求
type UpdateProviderRequest struct {
	Name         string  `json:"name" binding:"required,max=100" msg:"厂商名称不能为空"`
	Protocol     uint8   `json:"protocol" binding:"required,min=1,max=4" msg:"协议类型无效"`
	BaseURL      string  `json:"base_url"`
	AuthType     uint8   `json:"auth_type" binding:"max=2"`
	Capabilities []uint8 `json:"capabilities"`
}
