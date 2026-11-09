package common

import (
	"fmt"
	"os"
	"sync"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Hosts    map[string]HostConfig `yaml:"hosts"`
	Groups   map[string]HostConfig `yaml:"groups"`
	Loglevel string                `yaml:"loglevel"`
}

type SafeConfig struct {
	sync.RWMutex
	C *Config
}

type HostConfig struct {
	Username           string         `yaml:"username"`
	Password           string         `yaml:"password"`
	Collectlogs        bool           `yaml:"collectlogs"`
	Logcount           map[string]int `yaml:"logcount"`
	DisabledMetrics    []string       `yaml:"disabled_metrics"`
	ConcurrentRequests *int64         `yaml:"concurrent_requests"`
	AuthBackoffSecs    *int64         `yaml:"auth_backoff_secs"`
}

func (sc *SafeConfig) ReloadConfig(configFile string) error {
	var c = &Config{}

	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(yamlFile, c); err != nil {
		return err
	}

	for key, config := range c.Groups {
		config.ApplyDefaults()
		c.Groups[key] = config
	}

	for key, config := range c.Hosts {
		config.ApplyDefaults()
		c.Hosts[key] = config
	}

	sc.Lock()
	sc.C = c
	sc.Unlock()

	return nil
}

func (c *HostConfig) ApplyDefaults() {
	if c.DisabledMetrics == nil {
		c.DisabledMetrics = []string{}
	}
	if c.ConcurrentRequests == nil {
		c.ConcurrentRequests = new(int64)
		*c.ConcurrentRequests = 128
	}
	if c.AuthBackoffSecs == nil {
		c.AuthBackoffSecs = new(int64)
		*c.AuthBackoffSecs = 3600
	}
}

func (sc *SafeConfig) HostConfigForTarget(target string) (*HostConfig, error) {
	sc.Lock()
	defer sc.Unlock()
	if hostConfig, ok := sc.C.Hosts[target]; ok {
		return &hostConfig, nil
	}
	if hostConfig, ok := sc.C.Hosts["default"]; ok {
		return &hostConfig, nil
	}
	return &HostConfig{}, fmt.Errorf("no credentials found for target %s", target)
}

// HostConfigForGroup checks the configuration for a matching group config and returns the configured HostConfig for
// that matched group.
func (sc *SafeConfig) HostConfigForGroup(group string) (*HostConfig, error) {
	sc.Lock()
	defer sc.Unlock()
	if hostConfig, ok := sc.C.Groups[group]; ok {
		return &HostConfig{
			Username:    hostConfig.Username,
			Password:    hostConfig.Password,
			Collectlogs: hostConfig.Collectlogs,
			Logcount:    hostConfig.Logcount,
		}, nil
	}
	return &HostConfig{}, fmt.Errorf("no credentials found for group %s", group)
}

func (sc *SafeConfig) AppLogLevel() string {
	sc.Lock()
	defer sc.Unlock()
	logLevel := sc.C.Loglevel
	if logLevel != "" {
		return logLevel
	}
	return "info"
}
