package config

import (
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"gopkg.in/yaml.v3"
)

type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

type ReverseProxyConfig struct {
	Router  RouterSettings   `yaml:"router"`
	Limiter *LimiterSettings `yaml:"limiter,omitempty"`
}

type EdgeLimiterConfig struct {
	Limiter  LimiterSettings `yaml:"limiter"`
	IsGlobal *bool           `yaml:"is_global,omitempty"`
}

type MetricsConfig struct {
	Hosts []string `yaml:"hosts"`
}

type FileConfig struct {
	Proxy       ReverseProxyConfig `yaml:"proxy"`
	EdgeLimiter EdgeLimiterConfig  `yaml:"edge_limiter"`
	Metrics     MetricsConfig      `yaml:"metrics"`
}

type ServerConfig struct {
	Host         string        `env:"HOST"`
	Port         int           `env:"PORT"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT"`
}

type EnvConfig struct {
	EdgeLimiterRedisURL     string    `env:"EDGE_LIMITER_REDIS_URL"`
	InternalLimiterRedisURL string    `env:"INTERNAL_LIMITER_REDIS_URL"`
	LogLevel                *LogLevel `env:"LOG_LEVEL"`
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
