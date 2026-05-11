package valueobject

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCapability_String(t *testing.T) {
	tests := []struct {
		cap      Capability
		expected string
	}{
		{CapabilityChat, "chat"},
		{CapabilityEmbedding, "embedding"},
		{CapabilityRerank, "rerank"},
		{CapabilityTTS, "tts"},
		{CapabilitySTT, "stt"},
		{Capability(99), "unknown"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.cap.String())
	}
}

func TestCapability_Validate(t *testing.T) {
	assert.NoError(t, CapabilityChat.Validate())
	assert.NoError(t, CapabilityEmbedding.Validate())
	assert.NoError(t, CapabilityRerank.Validate())
	assert.NoError(t, CapabilityTTS.Validate())
	assert.NoError(t, CapabilitySTT.Validate())
	assert.Error(t, Capability(99).Validate())
}
