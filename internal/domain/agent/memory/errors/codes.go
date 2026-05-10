package errors

import domainErrors "github.com/dysodeng/app/internal/domain/shared/errors"

const (
	CodeMemoryNotFound    = "MEMORY_NOT_FOUND"
	CodeMemorySaveFailed  = "MEMORY_SAVE_FAILED"
	CodeMemoryQueryFailed = "MEMORY_QUERY_FAILED"
)

var (
	ErrMemoryNotFound    = domainErrors.NewMemoryError(CodeMemoryNotFound, "记忆不存在", nil)
	ErrMemorySaveFailed  = domainErrors.NewMemoryError(CodeMemorySaveFailed, "记忆保存失败", nil)
	ErrMemoryQueryFailed = domainErrors.NewMemoryError(CodeMemoryQueryFailed, "记忆查询失败", nil)
)
