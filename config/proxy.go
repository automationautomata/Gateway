package config

import (
	"time"

	"gopkg.in/yaml.v3"
)

type UpstreamsAliases map[string]string

type Caches map[string]time.Duration

type UpstreamSettings struct {
	UpstreamAlias string  `yaml:"upstream"`
	Cache         *Caches `yaml:"cache"`
}

type UpstreamDefault struct {
	*UpstreamSettings
}

func (d *UpstreamDefault) UnmarshalYAML(node *yaml.Node) error {
	d.UpstreamSettings = &UpstreamSettings{}
	if node.Kind == yaml.ScalarNode {
		if err := node.Decode(&d.UpstreamAlias); err != nil {
			return err
		}
		return nil
	}
	return node.Decode(&d.UpstreamSettings)
}

type Path struct {
	Path             string `yaml:"path"`
	UpstreamSettings `yaml:",inline"`
}

type Route struct {
	Host    string           `yaml:"host"`
	Paths   []Path           `yaml:"pathes"`
	Default *UpstreamDefault `yaml:"default,omitempty"`
}

type RouterSettings struct {
	UpstreamsAliases UpstreamsAliases `yaml:"upstreams"`
	Routes           []Route          `yaml:"routes"`
	Default          *UpstreamDefault `yaml:"default"`
}
