package errors

import (
	domainErrors "github.com/dysodeng/app/internal/domain/shared/errors"
)

const (
	CodeWorkspaceNotFound           = "WORKSPACE_NOT_FOUND"
	CodeWorkspaceAlreadyExists      = "WORKSPACE_ALREADY_EXISTS"
	CodeWorkspaceDisabled           = "WORKSPACE_DISABLED"
	CodeWorkspaceQueryFailed        = "WORKSPACE_QUERY_FAILED"
	CodeWorkspaceSaveFailed         = "WORKSPACE_SAVE_FAILED"
	CodeWorkspaceNameEmpty          = "WORKSPACE_NAME_EMPTY"
	CodeWorkspaceMemberNotFound     = "WORKSPACE_MEMBER_NOT_FOUND"
	CodeWorkspaceMemberExists       = "WORKSPACE_MEMBER_EXISTS"
	CodeWorkspaceMemberQueryFailed  = "WORKSPACE_MEMBER_QUERY_FAILED"
	CodeWorkspaceMemberSaveFailed   = "WORKSPACE_MEMBER_SAVE_FAILED"
	CodeWorkspaceMemberDeleteFailed = "WORKSPACE_MEMBER_DELETE_FAILED"
)

var (
	ErrWorkspaceNotFound           = domainErrors.NewWorkspaceError(CodeWorkspaceNotFound, "工作空间不存在", nil)
	ErrWorkspaceAlreadyExists      = domainErrors.NewWorkspaceError(CodeWorkspaceAlreadyExists, "工作空间已存在", nil)
	ErrWorkspaceDisabled           = domainErrors.NewWorkspaceError(CodeWorkspaceDisabled, "工作空间已被禁用", nil)
	ErrWorkspaceQueryFailed        = domainErrors.NewWorkspaceError(CodeWorkspaceQueryFailed, "工作空间查询失败", nil)
	ErrWorkspaceSaveFailed         = domainErrors.NewWorkspaceError(CodeWorkspaceSaveFailed, "工作空间保存失败", nil)
	ErrWorkspaceNameEmpty          = domainErrors.NewWorkspaceError(CodeWorkspaceNameEmpty, "工作空间名称不能为空", nil)
	ErrWorkspaceMemberNotFound     = domainErrors.NewWorkspaceError(CodeWorkspaceMemberNotFound, "工作空间成员不存在", nil)
	ErrWorkspaceMemberExists       = domainErrors.NewWorkspaceError(CodeWorkspaceMemberExists, "该用户已是工作空间管理员", nil)
	ErrWorkspaceMemberQueryFailed  = domainErrors.NewWorkspaceError(CodeWorkspaceMemberQueryFailed, "工作空间成员查询失败", nil)
	ErrWorkspaceMemberSaveFailed   = domainErrors.NewWorkspaceError(CodeWorkspaceMemberSaveFailed, "工作空间成员保存失败", nil)
	ErrWorkspaceMemberDeleteFailed = domainErrors.NewWorkspaceError(CodeWorkspaceMemberDeleteFailed, "工作空间成员删除失败", nil)
)
