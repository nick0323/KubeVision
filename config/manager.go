package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nick0323/K8sVision/model"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Manager struct {
	config     *model.Config
	viper      *viper.Viper
	logger     *zap.Logger
	configFile string
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
			m.logger.Warn("Config file not found, using default config", zap.String("configFile", configFile))
		} else {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		m.logger.Info("Config file loaded successfully", zap.String("configFile", m.viper.ConfigFileUsed()))
	}

	if err := m.viper.Unmarshal(m.config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	m.applyEnvironmentOverrides()

	m.logger.Info("Configuration loaded",
		zap.String("server", m.config.GetServerAddress()),
		zap.String("logLevel", m.config.Log.Level),
	)

	return nil
}

func (m *Manager) GetConfig() *model.Config {
	return m.config
}

func (m *Manager) Close() error {
	return nil
}

func (m *Manager) UpdateLogger(newLogger *zap.Logger) {
	m.logger = newLogger
}

func (m *Manager) applyEnvironmentOverrides() {
	if port := os.Getenv("K8SVISION_SERVER_PORT"); port != "" {
		m.config.Server.Port = port
	}
	if host := os.Getenv("K8SVISION_SERVER_HOST"); host != "" {
		m.config.Server.Host = host
	}
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		m.config.Kubernetes.Kubeconfig = kubeconfig
	}
	if apiServer := os.Getenv("K8SVISION_KUBERNETES_APISERVER"); apiServer != "" {
		m.config.Kubernetes.APIServer = apiServer
	}
	if token := os.Getenv("K8SVISION_KUBERNETES_TOKEN"); token != "" {
		m.config.Kubernetes.Token = token
	}
	if secret := os.Getenv("K8SVISION_JWT_SECRET"); secret != "" {
		m.config.JWT.Secret = secret
	}
	if level := os.Getenv("K8SVISION_LOG_LEVEL"); level != "" {
		m.config.Log.Level = level
	}
	if username := os.Getenv("K8SVISION_AUTH_USERNAME"); username != "" {
		m.config.Auth.Username = username
	}
	if password := os.Getenv("K8SVISION_AUTH_PASSWORD"); password != "" {
		m.config.Auth.Password = password
	}
	if maxFail := os.Getenv("K8SVISION_AUTH_MAX_FAIL"); maxFail != "" {
		if val, err := strconv.Atoi(maxFail); err == nil {
			m.config.Auth.MaxLoginFail = val
		}
	}
	if lockMinutes := os.Getenv("K8SVISION_AUTH_LOCK_MINUTES"); lockMinutes != "" {
		if val, err := strconv.Atoi(lockMinutes); err == nil {
			m.config.Auth.LockDuration = time.Duration(val) * time.Minute
		}
	}
}

func (m *Manager) Set(key string, value interface{}) {
	m.viper.Set(key, value)

	parts := strings.Split(key, ".")
	if len(parts) == 2 {
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
}

func (m *Manager) GetJWTSecret() []byte {
	return []byte(m.config.JWT.Secret)
}

func (m *Manager) GetAuthConfig() *model.AuthConfig {
	return &m.config.Auth
}
