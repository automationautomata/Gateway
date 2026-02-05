package config

import "time"

type HostRules struct {
	Host    string            `yaml:"host"`
	Pathes  map[string]string `yaml:"pathes"`
	Default *string           `yaml:"default"`
}

type ReverseProxyRules struct {
	Hosts   []HostRules `yaml:"hosts"`
	Default string      `yaml:"default"`
}

type ReverseProxyConfig struct {
	Rules         ReverseProxyRules `yaml:"rules"`
	LimiterConfig *LimiterConfig    `yaml:"limiter"`
}

type ServerConfig struct {
	Host         string        `yaml:"hosts"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type EdgeLimiterConfig struct {
	LimiterConfig   LimiterConfig `yaml:"limiter"`
	IsGlobalLimiter *bool         `yaml:"is_global_limiter, omitempty"`
}

type MetricsConfig struct {
	Hosts []string `yaml:"hosts"`
}

type Config struct {
	Server      ServerConfig       `yaml:"server"`
	Proxy       ReverseProxyConfig `yaml:"proxy"`
	EdgeLimiter EdgeLimiterConfig  `yaml:"edge_limiter"`
	Metrics     MetricsConfig      `yaml:"metrics"`
}
