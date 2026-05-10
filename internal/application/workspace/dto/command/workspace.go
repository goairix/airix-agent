package command

// CreateWorkspaceCommand 创建工作空间命令
type CreateWorkspaceCommand struct {
	Name        string
	Description string
	CreatedBy   string // UUID 字符串
}

// AssignAdminCommand 分配管理员命令
type AssignAdminCommand struct {
	WorkspaceID string
	UserID      string
}
