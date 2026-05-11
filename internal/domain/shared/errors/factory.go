package errors

// 领域名称常量
const (
	DomainCommon     = "common"
	DomainShared     = "shared"
	DomainUser       = "user"
	DomainFile       = "file"
	DomainPassport   = "passport"
	DomainPermission = "permission"
	DomainWorkspace  = "workspace"
	DomainAgent      = "agent"
	DomainSession    = "session"
	DomainMemory     = "memory"
)

// NewCommonError 创建通用领域错误
func NewCommonError(code, message string, err error) *DomainError {
	return NewDomainError(DomainCommon, code, message, err)
}

// NewShardError 创建共享领域错误
func NewShardError(code, message string, err error) *DomainError {
	return NewDomainError(DomainShared, code, message, err)
}

// NewUserError 创建用户领域错误
func NewUserError(code, message string, err error) *DomainError {
	return NewDomainError(DomainUser, code, message, err)
}

// NewFileError 创建文件领域错误
func NewFileError(code, message string, err error) *DomainError {
	return NewDomainError(DomainFile, code, message, err)
}

// NewPassportError 创建通行证领域错误
func NewPassportError(code, message string, err error) *DomainError {
	return NewDomainError(DomainPassport, code, message, err)
}

func NewPermissionError(code, message string, err error) *DomainError {
	return NewDomainError(DomainPermission, code, message, err)
}

// NewWorkspaceError 创建工作空间领域错误
func NewWorkspaceError(code, message string, err error) *DomainError {
	return NewDomainError(DomainWorkspace, code, message, err)
}

// NewAgentError 创建 Agent 领域错误
func NewAgentError(code, message string, err error) *DomainError {
	return NewDomainError(DomainAgent, code, message, err)
}

// NewSessionError 创建 Session 领域错误
func NewSessionError(code, message string, err error) *DomainError {
	return NewDomainError(DomainSession, code, message, err)
}

// NewMemoryError 创建 Memory 领域错误
func NewMemoryError(code, message string, err error) *DomainError {
	return NewDomainError(DomainMemory, code, message, err)
}

const DomainModel = "model"

// NewModelError 创建模型领域错误
func NewModelError(code, message string, err error) *DomainError {
	return NewDomainError(DomainModel, code, message, err)
}
