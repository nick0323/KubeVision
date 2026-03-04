package api

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/model"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type PasswordChangeRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

type PasswordGenerateRequest struct {
	Length int `json:"length,omitempty"`
}

type PasswordHashRequest struct {
	Password string `json:"password" binding:"required"`
}

type PasswordValidateRequest struct {
	Password       string `json:"password" binding:"required"`
	HashedPassword string `json:"hashedPassword" binding:"required"`
}

type PasswordManager struct{}

func NewPasswordManager() *PasswordManager {
	return &PasswordManager{}
}

func (pm *PasswordManager) HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("生成盐失败: %w", err)
	}

	passwordWithSalt := password + base64.URLEncoding.EncodeToString(salt)

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(passwordWithSalt), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密码哈希失败: %w", err)
	}

	return base64.URLEncoding.EncodeToString(salt) + ":" + string(hashedBytes), nil
}

func (pm *PasswordManager) VerifyPassword(password, hashedPassword string) bool {
	parts := strings.Split(hashedPassword, ":")
	if len(parts) != 2 {
		return false
	}

	salt, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}

	passwordWithSalt := password + base64.URLEncoding.EncodeToString(salt)

	err = bcrypt.CompareHashAndPassword([]byte(parts[1]), []byte(passwordWithSalt))
	return err == nil
}

func (pm *PasswordManager) GeneratePassword(length int) (string, error) {
	if length <= 0 {
		length = model.DefaultPasswordLen
	}

	if length > model.MaxPasswordLen {
		length = model.MaxPasswordLen
	}

	b := make([]byte, length)
	charsetBytes := []byte(model.PasswordCharset)

	for i := range b {
		randomIndex := make([]byte, 1)
		if _, err := rand.Read(randomIndex); err != nil {
			return "", fmt.Errorf("生成随机字符失败: %w", err)
		}
		b[i] = charsetBytes[randomIndex[0]%byte(len(charsetBytes))]
	}

	return string(b), nil
}

func (pm *PasswordManager) ValidatePasswordStrength(password string) (bool, string) {
	if len(password) < model.MinPasswordLen {
		return false, fmt.Sprintf("密码长度至少%d位", model.MinPasswordLen)
	}

	if len(password) > model.MaxPasswordLen {
		return false, fmt.Sprintf("密码长度不能超过%d位", model.MaxPasswordLen)
	}

	// 检查常见弱密码
	if pm.isWeakPassword(password) {
		return false, "密码过于简单，请使用更复杂的密码"
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return false, "密码必须包含大写字母"
	}
	if !hasLower {
		return false, "密码必须包含小写字母"
	}
	if !hasDigit {
		return false, "密码必须包含数字"
	}
	if !hasSpecial {
		return false, "密码必须包含特殊字符"
	}

	return true, "密码强度符合要求"
}

// isWeakPassword 检查是否为弱密码
func (pm *PasswordManager) isWeakPassword(password string) bool {
	weakPasswords := []string{
		"123456", "password", "admin", "root", "user", "test",
		"12345678", "qwerty", "abc123", "password123", "admin123",
		"123456789", "1234567890", "letmein", "welcome", "monkey",
		"dragon", "master", "hello", "login", "pass",
		"1234", "12345", "1234567", "123456789", "1234567890",
		"qwertyuiop", "asdfghjkl", "zxcvbnm", "password1",
		"admin1234", "root123", "user123", "test123",
	}

	passwordLower := strings.ToLower(password)
	for _, weak := range weakPasswords {
		if passwordLower == weak || strings.Contains(passwordLower, weak) {
			return true
		}
	}

	// 检查连续数字
	if pm.hasConsecutiveNumbers(password) {
		return true
	}

	// 检查重复字符
	if pm.hasRepeatedCharacters(password) {
		return true
	}

	return false
}

// hasConsecutiveNumbers 检查是否有连续数字
func (pm *PasswordManager) hasConsecutiveNumbers(password string) bool {
	consecutiveCount := 0
	for i := 0; i < len(password)-1; i++ {
		if password[i] >= '0' && password[i] <= '9' {
			if password[i+1] == password[i]+1 {
				consecutiveCount++
				if consecutiveCount >= 3 {
					return true
				}
			} else {
				consecutiveCount = 0
			}
		} else {
			consecutiveCount = 0
		}
	}
	return false
}

// hasRepeatedCharacters 检查是否有过多重复字符
func (pm *PasswordManager) hasRepeatedCharacters(password string) bool {
	charCount := make(map[rune]int)
	for _, char := range password {
		charCount[char]++
		if charCount[char] > len(password)/2 {
			return true
		}
	}
	return false
}

var passwordManager = NewPasswordManager()

func RegisterPasswordAdmin(r *gin.RouterGroup, logger *zap.Logger) {
	r.POST("/admin/password/change", changePassword(logger))
	r.POST("/admin/password/generate", generatePassword(logger))
	r.POST("/admin/password/hash", hashPassword(logger))
	r.POST("/admin/password/validate", validatePassword(logger))
}

// changePassword 修改密码
func changePassword(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PasswordChangeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.ResponseError(c, logger, &model.APIError{
				Code:    model.CodeBadRequest,
				Message: "请求参数格式错误",
				Details: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		// 验证新密码强度
		if valid, message := passwordManager.ValidatePasswordStrength(req.NewPassword); !valid {
			middleware.ResponseError(c, logger, &model.APIError{
				Code:    model.CodeValidationFailed,
				Message: "密码强度不符合要求",
				Details: message,
			}, http.StatusBadRequest)
			return
		}

		// 验证旧密码
		authConfig := configManager.GetAuthConfig()
		oldPasswordMatch := false

		if isHashedPassword(authConfig.Password) {
			oldPasswordMatch = passwordManager.VerifyPassword(req.OldPassword, authConfig.Password)
		} else {
			oldPasswordMatch = req.OldPassword == authConfig.Password
		}

		if !oldPasswordMatch {
			middleware.ResponseError(c, logger, &model.APIError{
				Code:    model.CodeAuthError,
				Message: "旧密码错误",
				Details: "请提供正确的旧密码",
			}, http.StatusBadRequest)
			return
		}

		// 检查新密码是否与旧密码相同
		if req.OldPassword == req.NewPassword {
			middleware.ResponseError(c, logger, &model.APIError{
				Code:    model.CodeValidationFailed,
				Message: "新密码不能与旧密码相同",
				Details: "请使用不同的新密码",
			}, http.StatusBadRequest)
			return
		}

		// 生成新密码的哈希值
		newHashedPassword, err := passwordManager.HashPassword(req.NewPassword)
		if err != nil {
			logger.Error("密码哈希失败", zap.Error(err))
			middleware.ResponseError(c, logger, &model.APIError{
				Code:    model.CodeInternalServerError,
				Message: "密码处理失败",
				Details: "无法生成密码哈希",
			}, http.StatusInternalServerError)
			return
		}

		// 持久化更新配置中的密码
		if configManager == nil {
			middleware.ResponseError(c, logger, &model.APIError{
				Code:    model.CodeInternalServerError,
				Message: "系统配置未初始化",
			}, http.StatusInternalServerError)
			return
		}

		// 更新 viper 中的配置并写回文件
		configManager.Set("auth.password", newHashedPassword)
		if err := configManager.WriteConfig(); err != nil {
			logger.Error("写入配置文件失败", zap.Error(err))
			middleware.ResponseError(c, logger, &model.APIError{
				Code:    model.CodeInternalServerError,
				Message: "配置持久化失败",
			}, http.StatusInternalServerError)
			return
		}

		// 记录审计日志（不暴露敏感信息）
		username := c.GetString("username")
		if username == "" {
			username = "admin"
		}
		logger.Info("密码修改成功", zap.String("username", username))

		middleware.ResponseSuccess(c, gin.H{
			"message": "密码修改成功",
		}, "密码修改成功", nil)
	}
}

// generatePassword 生成密码
func generatePassword(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PasswordGenerateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			// 使用默认长度
			req.Length = 12
		}

		password, err := passwordManager.GeneratePassword(req.Length)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		// 生成哈希值
		hashedPassword, err := passwordManager.HashPassword(password)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		// 默认不返回明文密码，仅返回长度与哈希；如需明文可通过前端受控开关另行支持
		middleware.ResponseSuccess(c, gin.H{
			"hashedPassword": hashedPassword,
			"length":         len(password),
		}, "密码生成成功", nil)
	}
}

// hashPassword 哈希密码
func hashPassword(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PasswordHashRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.ResponseError(c, logger, &model.APIError{
				Code:    model.CodeBadRequest,
				Message: "请求参数格式错误",
				Details: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		hashedPassword, err := passwordManager.HashPassword(req.Password)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, gin.H{
			"hashedPassword": hashedPassword,
		}, "密码哈希成功", nil)
	}
}

// validatePassword 验证密码
func validatePassword(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PasswordValidateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.ResponseError(c, logger, &model.APIError{
				Code:    model.CodeBadRequest,
				Message: "请求参数格式错误",
				Details: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		isValid := passwordManager.VerifyPassword(req.Password, req.HashedPassword)

		middleware.ResponseSuccess(c, gin.H{
			"valid": isValid,
			"message": func() string {
				if isValid {
					return "密码验证通过"
				}
				return "密码验证失败"
			}(),
		}, "验证完成", nil)
	}
}
