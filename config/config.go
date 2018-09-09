package config

import (
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Version string `yaml:"version"`
	Server  Server `yaml:"server"`
}

type Server struct {
	Port             int    `yaml:"port"`
	HealthzPort      int    `yaml:"health_check_port"`
	HealthzPath      string `yaml:"health_check_path"`
	Timeout          string `yaml:"timeout"`
	ShutdownDuration string `yaml:"shutdown_duration"`
	TLS              TLS    `yaml:"tls"`
}

type TLS struct {
	Enabled bool   `yaml:"enabled"`
	Cert    string `yaml:"cert"`
	Key     string `yaml:"key"`
	CA      string `yaml:"ca"`
}

const (
	currentVersion = "v1.0.0"
)

func New(path string) (*Config, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	cfg := new(Config)
	err = yaml.NewDecoder(f).Decode(&cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func GetVersion() string {
	return currentVersion
}

func GetValue(cfg string) string {
	if checkPrefixAndSuffix(cfg, "_", "_") {
		return os.Getenv(strings.TrimPrefix(strings.TrimSuffix(cfg, "_"), "_"))
	}
	return cfg
}

func checkPrefixAndSuffix(str, pref, suf string) bool {
	return strings.HasPrefix(str, pref) && strings.HasSuffix(str, suf)
}
