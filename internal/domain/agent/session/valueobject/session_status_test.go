package valueobject_test

import (
	"testing"

	"github.com/dysodeng/app/internal/domain/agent/session/valueobject"
)

func TestSessionStatus_Validate(t *testing.T) {
	if err := valueobject.SessionStatusRunning.Validate(); err != nil {
		t.Errorf("valid status should pass: %v", err)
	}
	var invalid valueobject.SessionStatus = 99
	if err := invalid.Validate(); err == nil {
		t.Error("invalid status should fail")
	}
}

func TestMessageItemType_Validate(t *testing.T) {
	if err := valueobject.MessageItemTypeAssistant.Validate(); err != nil {
		t.Errorf("valid type should pass: %v", err)
	}
	var invalid valueobject.MessageItemType = 99
	if err := invalid.Validate(); err == nil {
		t.Error("invalid type should fail")
	}
}
