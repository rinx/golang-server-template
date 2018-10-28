/*
Package config stores all server application configuration, to read the configuration file from yaml,
and decode the configuration to a Config struct.
*/
package config

import (
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

const (
	// currentVersion represent the config file version
	currentVersion = "v1.0.0"
)

// Config represent a application configuration content (config.yaml).
// In K8s environment, this configuration is stored in K8s ConfigMap.
type Config struct {
	// Version represent configuration file version.
	Version string `yaml:"version"`

	// Server represent server and health check server configuration.
	Server Server `yaml:"server"`
}

// Server represent server and health check server configuration.
type Server struct {
	// GrpcPort represent grpc API server port.
	GrpcPort int `yaml:"grpc_port"`

	// GrpcWebPort represent grpc Web API server port.
	GrpcWebPort int `yaml:"grpc_web_port"`

	// RestPort represent http Rest API server port.
	RestPort int `yaml:"http_port"`

	// HealthzPort represent health check server port for K8s.
	HealthzPort int `yaml:"health_check_port"`

	// HealthzPath represent the server path (pattern) for health check server.
	HealthzPath string `yaml:"health_check_path"`

	// Timeout represent the server timeout value.
	Timeout string `yaml:"timeout"`

	// ShutdownDuration represent the parse duration before the server shutdown.
	ShutdownDuration string `yaml:"shutdown_duration"`

	// ProbeWaitTime represent the parse duration between health check server and server shutdown.
	ProbeWaitTime string `yaml:"probe_wait_time"`

	// TLS represent the TLS configuration for server.
	TLS TLS `yaml:"tls"`
}

// TLS represent the TLS configuration for server.
type TLS struct {
	// Enable represent the server enable TLS or not.
	Enabled bool `yaml:"enabled"`

	// CertKey represent the certificate environment variable key used to start server.
	CertKey string `yaml:"cert_key"`

	// KeyKey represent the private key environment variable key used to start server.
	KeyKey string `yaml:"key_key"`

	// CAKey represent the CA certificate environment variable key used to start server.
	CAKey string `yaml:"ca_key"`
}

// New returns *Config or error when decode the configuration file to actually *Config struct.
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

// GetVersion returns the current version of the server version.
func GetVersion() string {
	return currentVersion
}

// GetActualValue returns the environment variable value if the val has prefix and suffix "_", otherwise the val will directly return.
func GetActualValue(val string) string {
	if checkPrefixAndSuffix(val, "_", "_") {
		return os.Getenv(strings.TrimPrefix(strings.TrimSuffix(val, "_"), "_"))
	}
	return val
}

// checkPrefixAndSuffix checks if the str has prefix and suffix
func checkPrefixAndSuffix(str, pref, suf string) bool {
	return strings.HasPrefix(str, pref) && strings.HasSuffix(str, suf)
}
