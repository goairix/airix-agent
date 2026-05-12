package model

// CreateInstanceRequest 创建模型实例请求
type CreateInstanceRequest struct {
	ProviderID   string         `json:"provider_id" binding:"required" msg:"模型厂商ID不能为空"`
	ModelName    string         `json:"model_name" binding:"required,max=100" msg:"模型名称不能为空"`
	Capability   uint8          `json:"capability" binding:"required,min=1,max=5" msg:"模型能力类型无效"`
	APIKey       string         `json:"api_key"`
	Parameters   map[string]any `json:"parameters"`
	RateLimitRPM int            `json:"rate_limit_rpm"`
	RateLimitTPM int            `json:"rate_limit_tpm"`
}

// UpdateInstanceRequest 更新模型实例请求
type UpdateInstanceRequest struct {
	ModelName    string         `json:"model_name" binding:"required,max=100" msg:"模型名称不能为空"`
	Capability   uint8          `json:"capability" binding:"required,min=1,max=5" msg:"模型能力类型无效"`
	APIKey       string         `json:"api_key"`
	Parameters   map[string]any `json:"parameters"`
	RateLimitRPM int            `json:"rate_limit_rpm"`
	RateLimitTPM int            `json:"rate_limit_tpm"`
}

// ListInstanceRequest 模型实例列表请求
type ListInstanceRequest struct {
	Page     int `form:"page" binding:"required,min=1" msg:"页码不能为空"`
	PageSize int `form:"page_size" binding:"required,min=1,max=100" msg:"每页数量无效"`
}
