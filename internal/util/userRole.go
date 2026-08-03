package util

// 检查用户是否是管理员（保持向后兼容）.
func IsAdmin(
	userRole int16,
) bool {
	return userRole == 0
}
