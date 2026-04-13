package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// 默认 bcrypt 成本因子（推荐值 12，比默认的 10 更安全）
const DefaultBcryptCost = 12

// HashPassword 对密码进行哈希处理（使用推荐的成本因子）
func HashPassword(password string) (string, error) {
	return HashPasswordWithCost(password, DefaultBcryptCost)
}

// HashPasswordWithCost 使用指定成本因子对密码进行哈希处理
func HashPasswordWithCost(password string, cost int) (string, error) {
	// 验证成本因子范围（4-31）
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = DefaultBcryptCost
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyPassword 验证密码
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
