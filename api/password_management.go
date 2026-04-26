package api

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/config"
	"github.com/nick0323/K8sVision/model"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	BcryptCost          = 12
	PasswordHistorySize = 5
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

type PasswordManager struct {
	mu              sync.RWMutex
	passwordHistory []string
}

func NewPasswordManager() *PasswordManager {
	return &PasswordManager{passwordHistory: make([]string, 0, PasswordHistorySize)}
}

func (pm *PasswordManager) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("password hashing failed: %w", err)
	}
	return string(hashedBytes), nil
}

func (pm *PasswordManager) VerifyPassword(password, hashedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

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
			return "", fmt.Errorf("failed to generate random character: %w", err)
		}
		b[i] = charsetBytes[randomIndex.Int64()]
	}
	return string(b), nil
}

func (pm *PasswordManager) ValidatePasswordStrength(password string) (bool, string) {
	if len(password) < model.MinPasswordLen {
		return false, fmt.Sprintf("password must be at least %d characters long", model.MinPasswordLen)
	}
	if len(password) > model.MaxPasswordLen {
		return false, fmt.Sprintf("password length cannot exceed %d characters", model.MaxPasswordLen)
	}

	if pm.isWeakPassword(password) {
		return false, "password is too weak, please use a more complex password"
	}

	hasUpper, hasLower, hasDigit, hasSpecial := false, false, false, false
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
		return false, "password must contain at least 3 character types"
	}
	return true, "password strength is acceptable"
}

func (pm *PasswordManager) IsPasswordInHistory(password, hashedPassword string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, hist := range pm.passwordHistory {
		if pm.VerifyPassword(password, hist) || hist == hashedPassword {
			return true
		}
	}
	return false
}

func (pm *PasswordManager) AddToPasswordHistory(hashedPassword string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, hist := range pm.passwordHistory {
		if hist == hashedPassword {
			return
		}
	}

	if len(pm.passwordHistory) >= PasswordHistorySize {
		pm.passwordHistory = pm.passwordHistory[1:]
	}
	pm.passwordHistory = append(pm.passwordHistory, hashedPassword)
}

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

	if pm.hasConsecutiveNumbers(password) || pm.hasRepeatedCharacters(password) {
		return true
	}
	return false
}

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

func isHashedPassword(password string) bool {
	if len(password) < 60 {
		return false
	}
	return strings.HasPrefix(password, "$2a$") ||
		strings.HasPrefix(password, "$2b$") ||
		strings.HasPrefix(password, "$2y$")
}

func RegisterPasswordAdmin(r *gin.RouterGroup, configMgr *config.Manager, logger *zap.Logger) {
	handler := &PasswordHandler{
		passwordManager: NewPasswordManager(),
		configManager:   configMgr,
		logger:          logger,
	}
	handler.RegisterRoutes(r)
}

type PasswordHandler struct {
	passwordManager *PasswordManager
	configManager   *config.Manager
	logger          *zap.Logger
}

func (h *PasswordHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/admin/password/change", h.ChangePassword())
	r.POST("/admin/password/generate", h.GeneratePassword())
	r.POST("/admin/password/hash", h.HashPassword())
	r.POST("/admin/password/validate", h.ValidatePassword())
}

func (h *PasswordHandler) ChangePassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PasswordChangeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.ResponseError(c, h.logger, &model.APIError{
				Code:    http.StatusBadRequest,
				Message: "Invalid request parameter format",
				Details: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		if valid, msg := h.passwordManager.ValidatePasswordStrength(req.NewPassword); !valid {
			middleware.ResponseError(c, h.logger, &model.APIError{
				Code:    model.CodeValidationFailed,
				Message: "Password strength does not meet requirements",
				Details: msg,
			}, http.StatusBadRequest)
			return
		}

		authConfig := h.configManager.GetConfig().Auth
		oldPasswordMatch := false
		if isHashedPassword(authConfig.Password) {
			oldPasswordMatch = h.passwordManager.VerifyPassword(req.OldPassword, authConfig.Password)
		} else {
			oldPasswordMatch = req.OldPassword == authConfig.Password
		}

		if !oldPasswordMatch {
			middleware.ResponseError(c, h.logger, &model.APIError{
				Code:    model.CodeUnauthorized,
				Message: "Invalid old password",
			}, http.StatusUnauthorized)
			return
		}

		if req.OldPassword == req.NewPassword {
			middleware.ResponseError(c, h.logger, &model.APIError{
				Code:    model.CodeValidationFailed,
				Message: "New password cannot be the same as old password",
			}, http.StatusBadRequest)
			return
		}

		newHashedPassword, err := h.passwordManager.HashPassword(req.NewPassword)
		if err != nil {
			h.logger.Error("Failed to hash password", zap.Error(err))
			middleware.ResponseError(c, h.logger, &model.APIError{
				Code:    model.CodeInternalServerError,
				Message: "Password processing failed",
			}, http.StatusInternalServerError)
			return
		}

		if h.passwordManager.IsPasswordInHistory(req.NewPassword, newHashedPassword) {
			middleware.ResponseError(c, h.logger, &model.APIError{
				Code:    model.CodeValidationFailed,
				Message: "Cannot use recently used password",
			}, http.StatusBadRequest)
			return
		}

		currentUser := GetUsernameFromContext(c)
		if currentUser == "" {
			currentUser = "admin"
		}

		h.configManager.Set("auth.password", newHashedPassword)
		h.passwordManager.AddToPasswordHistory(newHashedPassword)

		h.logger.Info("Password changed successfully (in-memory only)",
			zap.String("username", currentUser),
			zap.String("clientIP", c.ClientIP()),
		)

		middleware.ResponseSuccess(c, gin.H{"message": "Password changed successfully"}, "Password changed successfully", nil)
	}
}

func (h *PasswordHandler) GeneratePassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PasswordGenerateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			req.Length = model.DefaultPasswordLen
		}

		password, err := h.passwordManager.GeneratePassword(req.Length)
		if err != nil {
			middleware.ResponseError(c, h.logger, err, http.StatusInternalServerError)
			return
		}

		hashedPassword, err := h.passwordManager.HashPassword(password)
		if err != nil {
			middleware.ResponseError(c, h.logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, gin.H{
			"password":       password,
			"hashedPassword": hashedPassword,
			"length":         len(password),
			"warning":        "Please save the plaintext password securely",
		}, "Password generated successfully", nil)
	}
}

func (h *PasswordHandler) HashPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PasswordHashRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.ResponseError(c, h.logger, &model.APIError{
				Code:    http.StatusBadRequest,
				Message: "Invalid request parameter format",
				Details: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		hashedPassword, err := h.passwordManager.HashPassword(req.Password)
		if err != nil {
			middleware.ResponseError(c, h.logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, gin.H{
			"hashedPassword": hashedPassword,
			"cost":           BcryptCost,
		}, "Password hashed successfully", nil)
	}
}

func (h *PasswordHandler) ValidatePassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PasswordValidateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.ResponseError(c, h.logger, &model.APIError{
				Code:    http.StatusBadRequest,
				Message: "Invalid request parameter format",
				Details: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		isValid := h.passwordManager.VerifyPassword(req.Password, req.HashedPassword)
		message := "Password verification failed"
		if isValid {
			message = "Password verification passed"
		}

		middleware.ResponseSuccess(c, gin.H{
			"valid":   isValid,
			"message": message,
		}, "Verification completed", nil)
	}
}
