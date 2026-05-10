package workspace

// CreateWorkspaceRequest 创建工作空间请求
type CreateWorkspaceRequest struct {
	Name        string `json:"name" binding:"required" msg:"请输入工作空间名称"`
	Description string `json:"description"`
	CreatedBy   string `json:"created_by" binding:"required" msg:"请输入创建人ID"`
}

// AssignAdminRequest 分配管理员请求
type AssignAdminRequest struct {
	UserID string `json:"user_id" binding:"required" msg:"请输入用户ID"`
}
