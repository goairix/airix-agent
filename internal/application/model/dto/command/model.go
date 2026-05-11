package command

// CreateProviderCommand 创建模型厂商命令
type CreateProviderCommand struct {
	Name         string  `json:"name" binding:"required,max=100" msg:"厂商名称不能为空"`
	Protocol     uint8   `json:"protocol" binding:"required,min=1,max=4" msg:"协议类型无效"`
	BaseURL      string  `json:"base_url"`
	AuthType     uint8   `json:"auth_type" binding:"max=2"`
	Capabilities []uint8 `json:"capabilities"`
}

// UpdateProviderCommand 更新模型厂商命令
type UpdateProviderCommand struct {
	ProviderID   string  `json:"provider_id" binding:"required"`
	Name         string  `json:"name" binding:"required,max=100" msg:"厂商名称不能为空"`
	Protocol     uint8   `json:"protocol" binding:"required,min=1,max=4" msg:"协议类型无效"`
	BaseURL      string  `json:"base_url"`
	AuthType     uint8   `json:"auth_type" binding:"max=2"`
	Capabilities []uint8 `json:"capabilities"`
}

// CreateInstanceCommand 创建模型实例命令
type CreateInstanceCommand struct {
	WorkspaceID  string         `json:"workspace_id" binding:"required" msg:"工作空间ID不能为空"`
	ProviderID   string         `json:"provider_id" binding:"required" msg:"模型厂商ID不能为空"`
	ModelName    string         `json:"model_name" binding:"required,max=100" msg:"模型名称不能为空"`
	Capability   uint8          `json:"capability" binding:"required,min=1,max=5" msg:"模型能力类型无效"`
	APIKey       string         `json:"api_key"`
	Parameters   map[string]any `json:"parameters"`
	RateLimitRPM int            `json:"rate_limit_rpm"`
	RateLimitTPM int            `json:"rate_limit_tpm"`
}

// UpdateInstanceCommand 更新模型实例命令
type UpdateInstanceCommand struct {
	InstanceID   string         `json:"instance_id" binding:"required"`
	ModelName    string         `json:"model_name" binding:"required,max=100" msg:"模型名称不能为空"`
	Capability   uint8          `json:"capability" binding:"required,min=1,max=5" msg:"模型能力类型无效"`
	APIKey       string         `json:"api_key"`
	Parameters   map[string]any `json:"parameters"`
	RateLimitRPM int            `json:"rate_limit_rpm"`
	RateLimitTPM int            `json:"rate_limit_tpm"`
}
