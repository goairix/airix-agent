package response

// WorkspaceResponse 工作空间响应
type WorkspaceResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      uint8  `json:"status"`
	StatusText  string `json:"status_text"`
	CreatedBy   string `json:"created_by"`
	CreatedAt   string `json:"created_at"`
}

// WorkspaceListResponse 工作空间列表响应
type WorkspaceListResponse struct {
	Record []WorkspaceResponse `json:"record"`
	Total  int                 `json:"total"`
}

// MemberResponse 成员响应
type MemberResponse struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	UserID      string `json:"user_id"`
	Role        uint8  `json:"role"`
	RoleText    string `json:"role_text"`
	AssignedAt  string `json:"assigned_at"`
}

// MemberListResponse 成员列表响应
type MemberListResponse struct {
	Record []MemberResponse `json:"record"`
	Total  int              `json:"total"`
}
