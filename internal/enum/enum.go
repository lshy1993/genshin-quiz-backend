package enum

type TokenType string

//nolint:gosec // 常量名包含 Token 触发了 gosec G101 误报
const (
	TokenTypePasswordReset TokenType = "password_reset"
	TokenTypeEmailVerify   TokenType = "email_verification"
	TokenTypeMagicLink     TokenType = "magic_link"
)

func (t TokenType) String() string {
	return string(t)
}
