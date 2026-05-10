package response

import "time"

// AgentResponse Agent 基础信息响应
type AgentResponse struct {
	AgentID             string                      `json:"agent_id"`
	WorkspaceID         string                      `json:"workspace_id"`
	Name                string                      `json:"name"`
	Description         string                      `json:"description"`
	AgentType           string                      `json:"agent_type"`
	SystemPrompt        string                      `json:"system_prompt"`
	ModelConfig         ModelConfigResponse         `json:"model_config"`
	ToolBindings        []string                    `json:"tool_bindings"`
	KnowledgeBindings   []string                    `json:"knowledge_bindings"`
	SkillBindings       []string                    `json:"skill_bindings"`
	MCPBindings         []string                    `json:"mcp_bindings"`
	MemoryConfig        MemoryConfigResponse        `json:"memory_config"`
	CollaborationConfig CollaborationConfigResponse `json:"collaboration_config"`
	SandboxConfig       SandboxConfigResponse       `json:"sandbox_config"`
	ActiveReleaseID     string                      `json:"active_release_id"`
	Status              string                      `json:"status"`
	CreatedAt           time.Time                   `json:"created_at"`
}

// ModelConfigResponse 模型配置响应
type ModelConfigResponse struct {
	ModelInstanceID string         `json:"model_instance_id"`
	Parameters      map[string]any `json:"parameters"`
}

// SummarizationConfigResponse 摘要配置响应
type SummarizationConfigResponse struct {
	SummaryModelInstanceID string `json:"summary_model_instance_id"`
	TriggerTokenThreshold  int    `json:"trigger_token_threshold"`
}

// MemoryConfigResponse 记忆配置响应
type MemoryConfigResponse struct {
	MemoryDriverType    string                      `json:"memory_driver_type"`
	ContextMode         string                      `json:"context_mode"`
	ContextWindowSize   int                         `json:"context_window_size"`
	SummarizationConfig SummarizationConfigResponse `json:"summarization_config"`
	GlobalMemoryMode    string                      `json:"global_memory_mode"`
}

// CollaborationConfigResponse 协作配置响应
type CollaborationConfigResponse struct {
	SubAgentIDs        []string `json:"sub_agent_ids"`
	TransferPolicy     string   `json:"transfer_policy"`
	MaxDelegationDepth int      `json:"max_delegation_depth"`
}

// SandboxConfigResponse 沙盒配置响应
type SandboxConfigResponse struct {
	Enabled     bool   `json:"enabled"`
	SandboxType string `json:"sandbox_type"`
}

// AgentListResponse Agent 列表响应
type AgentListResponse struct {
	Record []AgentResponse `json:"record"`
	Total  int64           `json:"total"`
}

// AgentReleaseResponse Agent 版本响应
type AgentReleaseResponse struct {
	ReleaseID  string    `json:"release_id"`
	AgentID    string    `json:"agent_id"`
	ChangeLog  string    `json:"change_log"`
	Status     string    `json:"status"`
	ReleasedAt time.Time `json:"released_at"`
	ReleasedBy string    `json:"released_by"`
}

// AgentReleaseListResponse Agent 版本列表响应
type AgentReleaseListResponse struct {
	Record []AgentReleaseResponse `json:"record"`
	Total  int64                  `json:"total"`
}
