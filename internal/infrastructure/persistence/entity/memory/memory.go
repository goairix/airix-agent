package memory

import (
	"time"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/infrastructure/pkg/model"
)

// Memory 记忆数据实体
type Memory struct {
	model.DistributedPrimaryKeyID
	WorkspaceID uuid.UUID `gorm:"type:uuid;not null;index:memory_user_idx,priority:1;comment:工作空间ID"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index:memory_user_idx,priority:2;comment:用户ID"`
	MemoryType  uint8     `gorm:"type:tinyint(3);not null;index:memory_user_idx,priority:4;comment:记忆类型 1-session 2-global"`
	AgentID     uuid.UUID `gorm:"type:uuid;index:memory_agent_idx;comment:AgentID(session类型时有值，会话记忆按Agent+User区分)"`
	Content     string    `gorm:"type:text;not null;comment:记忆内容"`
	Tags        string    `gorm:"type:json;comment:标签JSON"`
	Importance  float64   `gorm:"type:decimal(4,2);not null;default:0;comment:重要性评分"`
	Date        time.Time `gorm:"type:date;not null;index:memory_user_idx,priority:3;comment:记忆日期"`
	model.Time
	model.SoftDelete
}

func (Memory) TableName() string { return "memories" }
