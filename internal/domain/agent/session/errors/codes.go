package errors

import domainErrors "github.com/dysodeng/app/internal/domain/shared/errors"

const (
	CodeSessionNotFound        = "SESSION_NOT_FOUND"
	CodeSessionAgentIDEmpty    = "SESSION_AGENT_ID_EMPTY"
	CodeSessionWorkspaceEmpty  = "SESSION_WORKSPACE_EMPTY"
	CodeSessionSaveFailed      = "SESSION_SAVE_FAILED"
	CodeSessionQueryFailed     = "SESSION_QUERY_FAILED"
	CodeMessageSaveFailed      = "MESSAGE_SAVE_FAILED"
	CodeMessageQueryFailed     = "MESSAGE_QUERY_FAILED"
	CodeMessageItemSaveFailed  = "MESSAGE_ITEM_SAVE_FAILED"
	CodeMessageItemQueryFailed = "MESSAGE_ITEM_QUERY_FAILED"
)

var (
	ErrSessionNotFound        = domainErrors.NewSessionError(CodeSessionNotFound, "会话不存在", nil)
	ErrSessionAgentIDEmpty    = domainErrors.NewSessionError(CodeSessionAgentIDEmpty, "Agent ID 不能为空", nil)
	ErrSessionWorkspaceEmpty  = domainErrors.NewSessionError(CodeSessionWorkspaceEmpty, "工作空间 ID 不能为空", nil)
	ErrSessionSaveFailed      = domainErrors.NewSessionError(CodeSessionSaveFailed, "会话保存失败", nil)
	ErrSessionQueryFailed     = domainErrors.NewSessionError(CodeSessionQueryFailed, "会话查询失败", nil)
	ErrMessageSaveFailed      = domainErrors.NewSessionError(CodeMessageSaveFailed, "消息保存失败", nil)
	ErrMessageQueryFailed     = domainErrors.NewSessionError(CodeMessageQueryFailed, "消息查询失败", nil)
	ErrMessageItemSaveFailed  = domainErrors.NewSessionError(CodeMessageItemSaveFailed, "消息步骤保存失败", nil)
	ErrMessageItemQueryFailed = domainErrors.NewSessionError(CodeMessageItemQueryFailed, "消息步骤查询失败", nil)
)
