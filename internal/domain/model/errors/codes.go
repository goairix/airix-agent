// internal/domain/model/errors/codes.go
package errors

import domainErrors "github.com/dysodeng/app/internal/domain/shared/errors"

// Provider 错误码
const (
	CodeProviderNotFound        = "MODEL_PROVIDER_NOT_FOUND"
	CodeProviderNameEmpty       = "MODEL_PROVIDER_NAME_EMPTY"
	CodeProviderProtocolInvalid = "MODEL_PROVIDER_PROTOCOL_INVALID"
	CodeProviderAuthTypeInvalid = "MODEL_PROVIDER_AUTH_TYPE_INVALID"
	CodeProviderSaveFailed      = "MODEL_PROVIDER_SAVE_FAILED"
	CodeProviderQueryFailed     = "MODEL_PROVIDER_QUERY_FAILED"
	CodeProviderDeleteFailed    = "MODEL_PROVIDER_DELETE_FAILED"
	CodeProviderHasInstances    = "MODEL_PROVIDER_HAS_INSTANCES"
)

// Instance 错误码
const (
	CodeInstanceNotFound          = "MODEL_INSTANCE_NOT_FOUND"
	CodeInstanceWorkspaceEmpty    = "MODEL_INSTANCE_WORKSPACE_EMPTY"
	CodeInstanceProviderEmpty     = "MODEL_INSTANCE_PROVIDER_EMPTY"
	CodeInstanceModelNameEmpty    = "MODEL_INSTANCE_MODEL_NAME_EMPTY"
	CodeInstanceCapabilityInvalid = "MODEL_INSTANCE_CAPABILITY_INVALID"
	CodeInstanceSaveFailed        = "MODEL_INSTANCE_SAVE_FAILED"
	CodeInstanceQueryFailed       = "MODEL_INSTANCE_QUERY_FAILED"
	CodeInstanceDeleteFailed      = "MODEL_INSTANCE_DELETE_FAILED"
	CodeInstanceDisabled          = "MODEL_INSTANCE_DISABLED"
)

// Provider 预定义错误
var (
	ErrProviderNotFound        = domainErrors.NewModelError(CodeProviderNotFound, "模型厂商不存在", nil)
	ErrProviderNameEmpty       = domainErrors.NewModelError(CodeProviderNameEmpty, "模型厂商名称不能为空", nil)
	ErrProviderProtocolInvalid = domainErrors.NewModelError(CodeProviderProtocolInvalid, "无效的模型协议类型", nil)
	ErrProviderAuthTypeInvalid = domainErrors.NewModelError(CodeProviderAuthTypeInvalid, "无效的认证类型", nil)
	ErrProviderSaveFailed      = domainErrors.NewModelError(CodeProviderSaveFailed, "模型厂商保存失败", nil)
	ErrProviderQueryFailed     = domainErrors.NewModelError(CodeProviderQueryFailed, "模型厂商查询失败", nil)
	ErrProviderDeleteFailed    = domainErrors.NewModelError(CodeProviderDeleteFailed, "模型厂商删除失败", nil)
	ErrProviderHasInstances    = domainErrors.NewModelError(CodeProviderHasInstances, "该厂商下仍有模型实例，无法删除", nil)
)

// Instance 预定义错误
var (
	ErrInstanceNotFound          = domainErrors.NewModelError(CodeInstanceNotFound, "模型实例不存在", nil)
	ErrInstanceWorkspaceEmpty    = domainErrors.NewModelError(CodeInstanceWorkspaceEmpty, "工作空间 ID 不能为空", nil)
	ErrInstanceProviderEmpty     = domainErrors.NewModelError(CodeInstanceProviderEmpty, "模型厂商 ID 不能为空", nil)
	ErrInstanceModelNameEmpty    = domainErrors.NewModelError(CodeInstanceModelNameEmpty, "模型名称不能为空", nil)
	ErrInstanceCapabilityInvalid = domainErrors.NewModelError(CodeInstanceCapabilityInvalid, "无效的模型能力类型", nil)
	ErrInstanceSaveFailed        = domainErrors.NewModelError(CodeInstanceSaveFailed, "模型实例保存失败", nil)
	ErrInstanceQueryFailed       = domainErrors.NewModelError(CodeInstanceQueryFailed, "模型实例查询失败", nil)
	ErrInstanceDeleteFailed      = domainErrors.NewModelError(CodeInstanceDeleteFailed, "模型实例删除失败", nil)
	ErrInstanceDisabled          = domainErrors.NewModelError(CodeInstanceDisabled, "模型实例已禁用", nil)
)
