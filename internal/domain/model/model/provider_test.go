// internal/domain/model/model/provider_test.go
package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dysodeng/app/internal/domain/model/valueobject"
)

func TestNewProvider(t *testing.T) {
	p, err := NewProvider("OpenAI", valueobject.ProtocolOpenAICompatible, "https://api.openai.com/v1", valueobject.AuthTypeAPIKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, p.ID)
	assert.Equal(t, "OpenAI", p.Name)
	assert.Equal(t, valueobject.ProtocolOpenAICompatible, p.Protocol)
	assert.Equal(t, "https://api.openai.com/v1", p.BaseURL)
	assert.Equal(t, valueobject.AuthTypeAPIKey, p.AuthType)
}

func TestNewProvider_NameEmpty(t *testing.T) {
	_, err := NewProvider("", valueobject.ProtocolOpenAICompatible, "https://api.openai.com/v1", valueobject.AuthTypeAPIKey)
	assert.Error(t, err)
}

func TestNewProvider_InvalidProtocol(t *testing.T) {
	_, err := NewProvider("OpenAI", valueobject.Protocol(99), "https://api.openai.com/v1", valueobject.AuthTypeAPIKey)
	assert.Error(t, err)
}

func TestProvider_AddCapability(t *testing.T) {
	p, _ := NewProvider("OpenAI", valueobject.ProtocolOpenAICompatible, "https://api.openai.com/v1", valueobject.AuthTypeAPIKey)
	p.AddCapability(valueobject.CapabilityChat)
	p.AddCapability(valueobject.CapabilityEmbedding)
	p.AddCapability(valueobject.CapabilityChat) // 重复添加应忽略
	assert.Len(t, p.SupportedCapabilities, 2)
}

func TestProvider_SupportsCapability(t *testing.T) {
	p, _ := NewProvider("OpenAI", valueobject.ProtocolOpenAICompatible, "https://api.openai.com/v1", valueobject.AuthTypeAPIKey)
	p.AddCapability(valueobject.CapabilityChat)
	assert.True(t, p.SupportsCapability(valueobject.CapabilityChat))
	assert.False(t, p.SupportsCapability(valueobject.CapabilityEmbedding))
}
