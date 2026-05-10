package model

import (
	"time"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/domain/memory/valueobject"
)

// Memory 记忆实体
type Memory struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	MemoryType  valueobject.MemoryType
	AgentID     uuid.UUID // session 类型时有值（会话记忆按 Agent+User 区分，跨会话持久）
	Content     string    // 结构化摘要或自然语言片段
	Tags        []string  // 检索标签
	Importance  float64   // 重要性评分，影响检索排序
	Date        time.Time // 按日期组织
	CreatedAt   time.Time
}

func NewMemory(workspaceID, userID uuid.UUID, memoryType valueobject.MemoryType, content string, tags []string, importance float64) *Memory {
	id, _ := uuid.NewV7()
	return &Memory{
		ID:          id,
		WorkspaceID: workspaceID,
		UserID:      userID,
		MemoryType:  memoryType,
		Content:     content,
		Tags:        tags,
		Importance:  importance,
		Date:        time.Now(),
		CreatedAt:   time.Now(),
	}
}
