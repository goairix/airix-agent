package valueobject_test

import (
	"testing"

	"github.com/dysodeng/app/internal/domain/workspace/valueobject"
)

func TestStatus_Validate(t *testing.T) {
	tests := []struct {
		name    string
		status  valueobject.Status
		wantErr bool
	}{
		{"active 有效", valueobject.StatusActive, false},
		{"disabled 有效", valueobject.StatusDisabled, false},
		{"非法值", valueobject.Status(9), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.status.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestStatus_IsActive(t *testing.T) {
	if !valueobject.StatusActive.IsActive() {
		t.Error("StatusActive.IsActive() = false, want true")
	}
	if valueobject.StatusDisabled.IsActive() {
		t.Error("StatusDisabled.IsActive() = true, want false")
	}
}

func TestStatus_String(t *testing.T) {
	if valueobject.StatusActive.String() != "active" {
		t.Errorf("StatusActive.String() = %s, want active", valueobject.StatusActive.String())
	}
	if valueobject.StatusDisabled.String() != "disabled" {
		t.Errorf("StatusDisabled.String() = %s, want disabled", valueobject.StatusDisabled.String())
	}
}
