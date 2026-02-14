package config

type Upstreams map[string]string

type Path struct {
	Path     string `yaml:"path"`
	Upstream string `yaml:"upstream"`
}

type Route struct {
	Host    string  `yaml:"host"`
	Paths   []Path  `yaml:"pathes"`
	Default *string `yaml:"default"`
}

type ProxySettings struct {
	Upstreams       Upstreams `yaml:"upstreams"`
	Routes          []Route   `yaml:"routes"`
	DefaultUpstream *string   `yaml:"default"`
}
