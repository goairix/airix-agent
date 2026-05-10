package valueobject

import "errors"

// MemberRole 工作空间成员角色
type MemberRole uint8

const (
	RoleSuperAdmin MemberRole = 1
	RoleAdmin      MemberRole = 2
)

func (r MemberRole) Uint8() uint8 {
	return uint8(r)
}

func (r MemberRole) String() string {
	switch r {
	case RoleSuperAdmin:
		return "super_admin"
	case RoleAdmin:
		return "admin"
	default:
		return "unknown"
	}
}

func (r MemberRole) IsSuperAdmin() bool {
	return r == RoleSuperAdmin
}

func (r MemberRole) Validate() error {
	if r != RoleSuperAdmin && r != RoleAdmin {
		return errors.New("无效的成员角色")
	}
	return nil
}
