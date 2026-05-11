package response

import "time"

// ProviderResponse 模型厂商响应
type ProviderResponse struct {
	ProviderID   string    `json:"provider_id"`
	Name         string    `json:"name"`
	Protocol     string    `json:"protocol"`
	BaseURL      string    `json:"base_url"`
	AuthType     string    `json:"auth_type"`
	Capabilities []string  `json:"capabilities"`
	CreatedAt    time.Time `json:"created_at"`
}

// ProviderListResponse 模型厂商列表响应
type ProviderListResponse struct {
	Record []ProviderResponse `json:"record"`
	Total  int64              `json:"total"`
}

// InstanceResponse 模型实例响应
type InstanceResponse struct {
	InstanceID   string         `json:"instance_id"`
	WorkspaceID  string         `json:"workspace_id"`
	ProviderID   string         `json:"provider_id"`
	ModelName    string         `json:"model_name"`
	Capability   string         `json:"capability"`
	Parameters   map[string]any `json:"parameters"`
	RateLimitRPM int            `json:"rate_limit_rpm"`
	RateLimitTPM int            `json:"rate_limit_tpm"`
	Status       string         `json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
}

// InstanceListResponse 模型实例列表响应
type InstanceListResponse struct {
	Record []InstanceResponse `json:"record"`
	Total  int64              `json:"total"`
}
