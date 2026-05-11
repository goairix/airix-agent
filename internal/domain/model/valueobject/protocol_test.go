// internal/domain/model/valueobject/protocol_test.go
package valueobject

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtocol_String(t *testing.T) {
	tests := []struct {
		protocol Protocol
		expected string
	}{
		{ProtocolOpenAICompatible, "openai-compatible"},
		{ProtocolAnthropic, "anthropic"},
		{ProtocolGoogle, "google"},
		{ProtocolCustom, "custom"},
		{Protocol(99), "unknown"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.protocol.String())
	}
}

func TestProtocol_Validate(t *testing.T) {
	assert.NoError(t, ProtocolOpenAICompatible.Validate())
	assert.NoError(t, ProtocolAnthropic.Validate())
	assert.NoError(t, ProtocolGoogle.Validate())
	assert.NoError(t, ProtocolCustom.Validate())
	assert.Error(t, Protocol(99).Validate())
}
