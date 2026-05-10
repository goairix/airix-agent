package valueobject

import "errors"

// MemoryType 记忆类型
type MemoryType uint8

const (
	MemoryTypeSession MemoryType = 1 // 会话记忆
	MemoryTypeGlobal  MemoryType = 2 // 全局记忆
)

func (t MemoryType) Uint8() uint8 { return uint8(t) }

func (t MemoryType) String() string {
	switch t {
	case MemoryTypeSession:
		return "session"
	case MemoryTypeGlobal:
		return "global"
	default:
		return "unknown"
	}
}

func (t MemoryType) Validate() error {
	switch t {
	case MemoryTypeSession, MemoryTypeGlobal:
		return nil
	}
	return errors.New("无效的记忆类型")
}
