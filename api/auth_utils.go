package api

import (
	"strings"
)

// isHashedPassword 判断密码是否为 bcrypt 哈希格式
// bcrypt 哈希以 $2a$、$2b$ 或 $2y$ 开头，长度至少 60 字符
func isHashedPassword(password string) bool {
	if len(password) < 60 {
		return false
	}
	return strings.HasPrefix(password, "$2a$") ||
		strings.HasPrefix(password, "$2b$") ||
		strings.HasPrefix(password, "$2y$")
}
