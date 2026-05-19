package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/nick0323/K8sVision/model"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
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

	m.viper.BindEnv("jwt.secret")
	m.viper.BindEnv("auth.password")
	m.viper.BindEnv("auth.username")

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

func (m *Manager) GetConfig() model.Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return *m.config
}

func (m *Manager) UpdateConfig(fn func(*model.Config)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	fn(m.config)
}

func (m *Manager) Close() error {
	return nil
}

func (m *Manager) UpdateLogger(logger *zap.Logger) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger = logger
}

func (m *Manager) Set(key string, value any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.viper.Set(key, value)

	parts := strings.Split(key, ".")
	if len(parts) != 2 {
		m.logger.Warn("config.Set() invalid key format, expected 'section.field'", zap.String("key", key))
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

func (m *Manager) GetClusters() []model.ClusterConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]model.ClusterConfig, len(m.config.Clusters))
	copy(result, m.config.Clusters)
	return result
}

func (m *Manager) AddCluster(cluster model.ClusterConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, c := range m.config.Clusters {
		if c.Name == cluster.Name {
			m.config.Clusters[i] = cluster
			return
		}
	}
	m.config.Clusters = append(m.config.Clusters, cluster)
}

func (m *Manager) RemoveCluster(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, c := range m.config.Clusters {
		if c.Name == name {
			m.config.Clusters = append(m.config.Clusters[:i], m.config.Clusters[i+1:]...)
			return
		}
	}
}

func (m *Manager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := yaml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	filename := m.configFile
	if filename == "" {
		filename = "config.yaml"
	}

	m.logger.Info("saving config", zap.String("file", filename))
	return os.WriteFile(filename, data, 0644)
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
