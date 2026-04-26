package config

import (
	"fmt"

	"go.uber.org/zap"
)

type SecurityChecker struct {
	configMgr *Manager
	logger    *zap.Logger
}

func NewSecurityChecker(configMgr *Manager, logger *zap.Logger) *SecurityChecker {
	return &SecurityChecker{configMgr: configMgr, logger: logger}
}

func (s *SecurityChecker) CheckAndValidate() error {
	cfg := s.configMgr.GetConfig()

	s.logger.Info("Validating security configuration",
		zap.Bool("jwt_secret_configured", cfg.JWT.Secret != ""),
		zap.Bool("auth_password_configured", cfg.Auth.Password != ""),
	)

	if cfg.JWT.Secret == "" {
		return fmt.Errorf("JWT Secret is not configured. Please set it via:\n" +
			"  1. config.yaml: jwt.secret: \"your-secret-key\"\n" +
			"  2. Environment variable: K8SVISION_JWT_SECRET=your-secret-key")
	}

	if len(cfg.JWT.Secret) < 32 {
		return fmt.Errorf("JWT Secret length must be at least 32 characters (current: %d)", len(cfg.JWT.Secret))
	}

	s.logger.Info("JWT Secret validation passed", zap.Int("length", len(cfg.JWT.Secret)))

	if cfg.Auth.Password == "" {
		return fmt.Errorf("Admin password is not configured. Please set it via:\n" +
			"  1. config.yaml: auth.password: \"your-password\"\n" +
			"  2. Environment variable: K8SVISION_AUTH_PASSWORD=your-password")
	}

	s.logger.Info("Admin password validation passed", zap.String("username", cfg.Auth.Username))
	s.logger.Info("Security configuration validation completed successfully")
	return nil
}
