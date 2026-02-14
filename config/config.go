package config

import (
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"gopkg.in/yaml.v3"
)

type LogLevel string

const (
	Debug LogLevel = "DEBUG"
	Info  LogLevel = "INFO"
	Warn  LogLevel = "WARN"
	Error LogLevel = "ERROR"
)

type ReverseProxyConfig struct {
	ProxySettings `yaml:",inline"`
	Limiter       *LimiterSettings `yaml:"limiter"`
}

type EdgeLimiterConfig struct {
	Limiter         LimiterSettings `yaml:"limiter"`
	IsGlobalLimiter *bool           `yaml:"is_global_limiter,omitempty"`
}

type MetricsConfig struct {
	Hosts []string `yaml:"hosts"`
}

type FileConfig struct {
	Proxy       ReverseProxyConfig `yaml:"proxy"`
	EdgeLimiter *EdgeLimiterConfig `yaml:"edge_limiter"`
	Metrics     MetricsConfig      `yaml:"metrics"`
}

type ServerConfig struct {
	Host         string        `env:"HOST"`
	Port         int           `env:"PORT"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT"`
}

type EnvConfig struct {
	EdgeLimiterRedisURL  string   `env:"EDGE_LIMITER_REDIS_URL"`
	ProxyLimiterRedisURL string   `env:"PROXY_LIMITER_REDIS_URL"`
	LogLevel             LogLevel `env:"LOG_LEVEL" envDefault:"ERROR"`
	ServerConfig
}

func LoadFileConfig(path string) (FileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FileConfig{}, err
	}

	var cfg FileConfig
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return FileConfig{}, err
	}
	return cfg, nil
}

func LoadEnvConfig(path string) (EnvConfig, error) {
	var cfg EnvConfig
	if err := env.Parse(&cfg); err != nil {
		return EnvConfig{}, err
	}
	return cfg, nil
}
