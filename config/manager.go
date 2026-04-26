package config

import (
	"fmt"
	"strings"
	"sync"

	"github.com/nick0323/K8sVision/model"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Manager struct {
	config     *model.Config
	viper      *viper.Viper
	logger     *zap.Logger
	configFile string
	mu         sync.RWMutex
}

func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		config: model.DefaultConfig(),
		viper:  viper.New(),
		logger: logger,
	}
}

func (m *Manager) Load(configFile string) error {
	m.configFile = configFile

	if configFile != "" {
		m.viper.SetConfigFile(configFile)
	} else {
		m.viper.SetConfigName("config")
		m.viper.AddConfigPath(".")
		m.viper.AddConfigPath("./config")
		m.viper.AddConfigPath("/etc/k8svision")
	}

	m.viper.SetEnvPrefix("K8SVISION")
	m.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	m.viper.AutomaticEnv()

	if err := m.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			m.logger.Warn("Config file not found, using default config")
		} else {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		m.logger.Info("Config file loaded", zap.String("file", m.viper.ConfigFileUsed()))
	}

	if err := m.viper.Unmarshal(m.config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	m.logger.Info("Configuration loaded",
		zap.String("server", m.config.GetServerAddress()),
		zap.String("logLevel", m.config.Log.Level),
	)

	return nil
}

func (m *Manager) GetConfig() *model.Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

func (m *Manager) Close() error {
	return nil
}

func (m *Manager) UpdateLogger(logger *zap.Logger) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger = logger
}

func (m *Manager) Set(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.viper.Set(key, value)

	parts := strings.Split(key, ".")
	if len(parts) != 2 {
		return
	}

	switch parts[0] {
	case "auth":
		switch parts[1] {
		case "password":
			if s, ok := value.(string); ok {
				m.config.Auth.Password = s
			}
		case "username":
			if s, ok := value.(string); ok {
				m.config.Auth.Username = s
			}
		}
	case "jwt":
		switch parts[1] {
		case "secret":
			if s, ok := value.(string); ok {
				m.config.JWT.Secret = s
			}
		}
	}
}

func (m *Manager) GetJWTSecret() []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return []byte(m.config.JWT.Secret)
}

func (m *Manager) GetAuthConfig() *model.AuthConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return &m.config.Auth
}
