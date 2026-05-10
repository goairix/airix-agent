package session

import (
	"time"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/infrastructure/pkg/model"
)

// Message 一轮对话数据实体
type Message struct {
	model.DistributedPrimaryKeyID
	SessionID           uuid.UUID  `gorm:"type:uuid;not null;index:message_session_sort_idx,priority:1;comment:会话ID"`
	WorkspaceID         uuid.UUID  `gorm:"type:uuid;not null;comment:工作空间ID"`
	AgentID             uuid.UUID  `gorm:"type:uuid;not null;comment:AgentID"`
	SortOrder           int64      `gorm:"type:bigint;not null;index:message_session_sort_idx,priority:2;comment:轮次序号(时间戳)"`
	Query               string     `gorm:"type:text;not null;comment:用户原始输入"`
	AgentInput          string     `gorm:"type:json;comment:Agent注入变量参数JSON"`
	Status              uint8      `gorm:"type:tinyint(3);not null;default:1;comment:状态 1-运行中 2-完成 3-失败 4-中断"`
	TotalTokens         int64      `gorm:"type:bigint;not null;default:0;comment:总token"`
	InputTokens         int64      `gorm:"type:bigint;not null;default:0;comment:输入token"`
	OutputTokens        int64      `gorm:"type:bigint;not null;default:0;comment:输出token"`
	CachedTokens        int64      `gorm:"type:bigint;not null;default:0;comment:缓存token"`
	ExecutionTimeMs     int64      `gorm:"type:bigint;not null;default:0;comment:整轮执行耗时(ms)"`
	FirstTokenLatencyMs int64      `gorm:"type:bigint;not null;default:0;comment:首token耗时(ms)"`
	CreatedAt           time.Time  `gorm:"type:timestamp(0) without time zone;not null;comment:创建时间"`
	CompletedAt         *time.Time `gorm:"type:timestamp(0) without time zone;comment:完成时间"`
}

func (Message) TableName() string { return "session_messages" }
