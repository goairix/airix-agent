package errors

import domainErrors "github.com/dysodeng/app/internal/domain/shared/errors"

const (
	CodeAgentNotFound        = "AGENT_NOT_FOUND"
	CodeAgentNameEmpty       = "AGENT_NAME_EMPTY"
	CodeAgentTypeInvalid     = "AGENT_TYPE_INVALID"
	CodeAgentWorkspaceEmpty  = "AGENT_WORKSPACE_EMPTY"
	CodeAgentSaveFailed      = "AGENT_SAVE_FAILED"
	CodeAgentQueryFailed     = "AGENT_QUERY_FAILED"
	CodeAgentDeleteFailed    = "AGENT_DELETE_FAILED"
	CodeAgentNoActiveRelease = "AGENT_NO_ACTIVE_RELEASE"

	CodeReleaseNotFound    = "AGENT_RELEASE_NOT_FOUND"
	CodeReleaseSaveFailed  = "AGENT_RELEASE_SAVE_FAILED"
	CodeReleaseQueryFailed = "AGENT_RELEASE_QUERY_FAILED"
)

var (
	ErrAgentNotFound        = domainErrors.NewAgentError(CodeAgentNotFound, "Agent 不存在", nil)
	ErrAgentNameEmpty       = domainErrors.NewAgentError(CodeAgentNameEmpty, "Agent 名称不能为空", nil)
	ErrAgentTypeInvalid     = domainErrors.NewAgentError(CodeAgentTypeInvalid, "无效的 Agent 类型", nil)
	ErrAgentWorkspaceEmpty  = domainErrors.NewAgentError(CodeAgentWorkspaceEmpty, "工作空间 ID 不能为空", nil)
	ErrAgentSaveFailed      = domainErrors.NewAgentError(CodeAgentSaveFailed, "Agent 保存失败", nil)
	ErrAgentQueryFailed     = domainErrors.NewAgentError(CodeAgentQueryFailed, "Agent 查询失败", nil)
	ErrAgentDeleteFailed    = domainErrors.NewAgentError(CodeAgentDeleteFailed, "Agent 删除失败", nil)
	ErrAgentNoActiveRelease = domainErrors.NewAgentError(CodeAgentNoActiveRelease, "Agent 尚未发布，不允许调用", nil)

	ErrReleaseNotFound    = domainErrors.NewAgentError(CodeReleaseNotFound, "Agent 版本不存在", nil)
	ErrReleaseSaveFailed  = domainErrors.NewAgentError(CodeReleaseSaveFailed, "Agent 版本保存失败", nil)
	ErrReleaseQueryFailed = domainErrors.NewAgentError(CodeReleaseQueryFailed, "Agent 版本查询失败", nil)
)
