package valueobject_test

import (
	"testing"

	"github.com/dysodeng/app/internal/domain/workspace/valueobject"
)

func TestMemberRole_Validate(t *testing.T) {
	tests := []struct {
		name    string
		role    valueobject.MemberRole
		wantErr bool
	}{
		{"超级管理员有效", valueobject.RoleSuperAdmin, false},
		{"管理员有效", valueobject.RoleAdmin, false},
		{"非法值", valueobject.MemberRole(9), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.role.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestMemberRole_IsSuperAdmin(t *testing.T) {
	if !valueobject.RoleSuperAdmin.IsSuperAdmin() {
		t.Error("RoleSuperAdmin.IsSuperAdmin() = false, want true")
	}
	if valueobject.RoleAdmin.IsSuperAdmin() {
		t.Error("RoleAdmin.IsSuperAdmin() = true, want false")
	}
}

func TestMemberRole_String(t *testing.T) {
	if valueobject.RoleSuperAdmin.String() != "super_admin" {
		t.Errorf("RoleSuperAdmin.String() = %s, want super_admin", valueobject.RoleSuperAdmin.String())
	}
	if valueobject.RoleAdmin.String() != "admin" {
		t.Errorf("RoleAdmin.String() = %s, want admin", valueobject.RoleAdmin.String())
	}
}
