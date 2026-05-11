package valueobject

import "errors"

// Capability 模型能力类型
type Capability uint8

const (
	CapabilityChat      Capability = 1
	CapabilityEmbedding Capability = 2
	CapabilityRerank    Capability = 3
	CapabilityTTS       Capability = 4
	CapabilitySTT       Capability = 5
)

func (c Capability) Uint8() uint8 {
	return uint8(c)
}

func (c Capability) String() string {
	switch c {
	case CapabilityChat:
		return "chat"
	case CapabilityEmbedding:
		return "embedding"
	case CapabilityRerank:
		return "rerank"
	case CapabilityTTS:
		return "tts"
	case CapabilitySTT:
		return "stt"
	default:
		return "unknown"
	}
}

func (c Capability) Validate() error {
	switch c {
	case CapabilityChat, CapabilityEmbedding, CapabilityRerank, CapabilityTTS, CapabilitySTT:
		return nil
	}
	return errors.New("无效的模型能力类型")
}
