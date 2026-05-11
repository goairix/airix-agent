// internal/domain/model/valueobject/protocol.go
package valueobject

import "errors"

// Protocol 模型协议类型
type Protocol uint8

const (
	ProtocolOpenAICompatible Protocol = 1
	ProtocolAnthropic        Protocol = 2
	ProtocolGoogle           Protocol = 3
	ProtocolCustom           Protocol = 4
)

func (p Protocol) Uint8() uint8 {
	return uint8(p)
}

func (p Protocol) String() string {
	switch p {
	case ProtocolOpenAICompatible:
		return "openai-compatible"
	case ProtocolAnthropic:
		return "anthropic"
	case ProtocolGoogle:
		return "google"
	case ProtocolCustom:
		return "custom"
	default:
		return "unknown"
	}
}

func (p Protocol) Validate() error {
	switch p {
	case ProtocolOpenAICompatible, ProtocolAnthropic, ProtocolGoogle, ProtocolCustom:
		return nil
	}
	return errors.New("无效的模型协议类型")
}
