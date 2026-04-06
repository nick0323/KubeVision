package api

import (
	"encoding/base64"
	"strings"
)

// parseHashedPassword 解析哈希密码格式（兼容旧格式和新格式）
// 旧格式：base64(salt):bcrypt_hash
// 新格式：bcrypt_hash（标准）
// 返回：salt(原始字节), hashPart(bcrypt 哈希), ok(是否有效)
func parseHashedPassword(hashedPassword string) ([]byte, string, bool) {
	// 检查是否为旧格式：base64(salt):bcrypt_hash
	if strings.Contains(hashedPassword, ":") {
		parts := strings.SplitN(hashedPassword, ":", 2)
		if len(parts) != 2 {
			return nil, "", false
		}

		// 解码 salt
		salt, err := base64.URLEncoding.DecodeString(parts[0])
		if err != nil {
			return nil, "", false
		}

		hashPart := parts[1]

		// 验证 bcrypt 哈希格式
		if !isValidBcryptHash(hashPart) {
			return nil, "", false
		}

		return salt, hashPart, true
	}

	// 新格式：标准 bcrypt 哈希（以 $2a$、$2b$ 或 $2y$ 开头）
	if isValidBcryptHash(hashedPassword) {
		return nil, hashedPassword, true
	}

	return nil, "", false
}

// isValidBcryptHash 验证是否为有效的 bcrypt 哈希
func isValidBcryptHash(hash string) bool {
	// bcrypt 哈希标准长度为 60 字符，但可能更长
	if len(hash) < 60 {
		return false
	}

	// 检查 bcrypt 版本前缀
	return strings.HasPrefix(hash, "$2a$") ||
		strings.HasPrefix(hash, "$2b$") ||
		strings.HasPrefix(hash, "$2y$")
}

// isHashedPassword 判断密码是否为哈希格式（兼容旧格式和新格式）
func isHashedPassword(password string) bool {
	_, _, ok := parseHashedPassword(password)
	return ok
}
