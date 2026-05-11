package valueobject

import "errors"

// AuthType 认证类型
type AuthType uint8

const (
	AuthTypeNone   AuthType = 0
	AuthTypeAPIKey AuthType = 1
	AuthTypeOAuth  AuthType = 2
)

func (a AuthType) Uint8() uint8 {
	return uint8(a)
}

func (a AuthType) String() string {
	switch a {
	case AuthTypeNone:
		return "none"
	case AuthTypeAPIKey:
		return "api-key"
	case AuthTypeOAuth:
		return "oauth"
	default:
		return "unknown"
	}
}

func (a AuthType) Validate() error {
	switch a {
	case AuthTypeNone, AuthTypeAPIKey, AuthTypeOAuth:
		return nil
	}
	return errors.New("无效的认证类型")
}
