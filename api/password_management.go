package api

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/model"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// bcrypt cost 配置
const (
	BcryptCost          = 12 // 推荐值 10-14，越高越安全但越慢
	PasswordHistorySize = 5  // 密码历史记录数量
)

// PasswordChangeRequest 密码修改请求
type PasswordChangeRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

// PasswordGenerateRequest 密码生成请求
type PasswordGenerateRequest struct {
	Length int `json:"length,omitempty"`
}

// PasswordHashRequest 密码哈希请求
type PasswordHashRequest struct {
	Password string `json:"password" binding:"required"`
}

// PasswordValidateRequest 密码验证请求
type PasswordValidateRequest struct {
	Password       string `json:"password" binding:"required"`
	HashedPassword string `json:"hashedPassword" binding:"required"`
}

// PasswordManager 密码管理器
type PasswordManager struct {
	mu              sync.RWMutex
	passwordHistory []string // 密码历史记录（哈希值）
}

// NewPasswordManager 创建密码管理器
func NewPasswordManager() *PasswordManager {
	return &PasswordManager{
		passwordHistory: make([]string, 0, PasswordHistorySize),
	}
}

// HashPassword 密码哈希（使用标准 bcrypt，内置盐）
func (pm *PasswordManager) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("密码哈希失败：%w", err)
	}
	return string(hashedBytes), nil
}

// VerifyPassword 验证密码（使用标准 bcrypt 验证）
func (pm *PasswordManager) VerifyPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GeneratePassword 生成随机密码（使用加密安全的随机数）
func (pm *PasswordManager) GeneratePassword(length int) (string, error) {
	if length <= 0 {
		length = model.DefaultPasswordLen
	}
	if length > model.MaxPasswordLen {
		length = model.MaxPasswordLen
	}

	charsetBytes := []byte(model.PasswordCharset)
	charsetLen := big.NewInt(int64(len(charsetBytes)))

	b := make([]byte, length)
	for i := range b {
		randomIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("生成随机字符失败：%w", err)
		}
		b[i] = charsetBytes[randomIndex.Int64()]
	}

	return string(b), nil
}

// ValidatePasswordStrength 验证密码强度
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

	// 检查密码复杂度
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

	// 至少满足 3 种字符类型
	charTypes := 0
	if hasUpper {
		charTypes++
	}
	if hasLower {
		charTypes++
	}
	if hasDigit {
		charTypes++
	}
	if hasSpecial {
		charTypes++
	}

	if charTypes < 3 {
		return false, "密码必须包含至少 3 种字符类型（大写字母、小写字母、数字、特殊字符）"
	}

	return true, "密码强度符合要求"
}

// IsPasswordInHistory 检查密码是否在历史记录中
func (pm *PasswordManager) IsPasswordInHistory(password, hashedPassword string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// 检查明文
	for _, hist := range pm.passwordHistory {
		if pm.VerifyPassword(password, hist) {
			return true
		}
	}

	// 检查哈希值
	for _, hist := range pm.passwordHistory {
		if hist == hashedPassword {
			return true
		}
	}

	return false
}

// AddToPasswordHistory 添加密码到历史记录
func (pm *PasswordManager) AddToPasswordHistory(hashedPassword string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 避免重复
	for _, hist := range pm.passwordHistory {
		if hist == hashedPassword {
			return
		}
	}

	// 添加到历史记录
	if len(pm.passwordHistory) >= PasswordHistorySize {
		pm.passwordHistory = pm.passwordHistory[1:]
	}
	pm.passwordHistory = append(pm.passwordHistory, hashedPassword)
}

// isWeakPassword 检查是否为弱密码
func (pm *PasswordManager) isWeakPassword(password string) bool {
	weakPasswords := []string{
		"123456", "password", "admin", "root", "user", "test",
		"12345678", "qwerty", "abc123", "password123", "admin123",
		"letmein", "welcome", "monkey", "dragon", "master",
		"hello", "login", "pass", "1234", "12345",
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

// RegisterPasswordAdmin 注册密码管理路由
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
		authConfig := GetAuthConfig()
		oldPasswordMatch := false

		if isHashedPassword(authConfig.Password) {
			oldPasswordMatch = passwordManager.VerifyPassword(req.OldPassword, authConfig.Password)
		} else {
			oldPasswordMatch = req.OldPassword == authConfig.Password
		}

		if !oldPasswordMatch {
			middleware.ResponseError(c, logger, &model.APIError{
				Code:    model.CodeUnauthorized,
				Message: "旧密码错误",
				Details: "请提供正确的旧密码",
			}, http.StatusUnauthorized)
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

		// 检查新密码是否在历史记录中
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

		if passwordManager.IsPasswordInHistory(req.NewPassword, newHashedPassword) {
			middleware.ResponseError(c, logger, &model.APIError{
				Code:    model.CodeValidationFailed,
				Message: "不能使用最近使用过的密码",
				Details: "请使用新的密码",
			}, http.StatusBadRequest)
			return
		}

		// 获取当前用户
		currentUser := GetUsernameFromContext(c)
		if currentUser == "" {
			currentUser = "admin"
		}

		// 持久化更新配置中的密码
		configManager.Set("auth.password", newHashedPassword)
		if err := configManager.WriteConfig(); err != nil {
			logger.Error("写入配置文件失败", zap.Error(err))
			middleware.ResponseError(c, logger, &model.APIError{
				Code:    model.CodeInternalServerError,
				Message: "配置持久化失败",
			}, http.StatusInternalServerError)
			return
		}

		// 添加到密码历史记录
		passwordManager.AddToPasswordHistory(newHashedPassword)

		// 记录审计日志（不暴露敏感信息）
		logger.Info("密码修改成功",
			zap.String("username", currentUser),
			zap.String("clientIP", c.ClientIP()),
			zap.Time("timestamp", time.Now()),
		)

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
			req.Length = model.DefaultPasswordLen
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

		// 返回明文密码和哈希值（生成场景特殊处理）
		middleware.ResponseSuccess(c, gin.H{
			"password":       password,
			"hashedPassword": hashedPassword,
			"length":         len(password),
			"warning":        "请安全保存明文密码，系统将不会再次显示",
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
			"cost":           BcryptCost,
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

		message := "密码验证失败"
		if isValid {
			message = "密码验证通过"
		}

		middleware.ResponseSuccess(c, gin.H{
			"valid":   isValid,
			"message": message,
		}, "验证完成", nil)
	}
}
