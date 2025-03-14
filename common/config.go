package common

import (
	"fmt"
	"os"
	"sync"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Hosts    map[string]HostConfigYaml `yaml:"hosts"`
	Groups   map[string]HostConfigYaml `yaml:"groups"`
	Loglevel string                    `yaml:"loglevel"`
}

type SafeConfig struct {
	sync.RWMutex
	C *Config
}

type HostConfigYaml struct {
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	Collectlogs *bool  `yaml:"collectlogs,omitempty"`
	Logcount    *int   `yaml:"logcount,omitempty"`
}

type HostConfig struct {
	Username    string
	Password    string
	Collectlogs bool
	Logcount    int
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

	sc.Lock()
	sc.C = c
	sc.Unlock()

	return nil
}

func (sc *SafeConfig) HostConfigForTarget(target string) (*HostConfig, error) {
	sc.Lock()
	defer sc.Unlock()
	if hostConfig, ok := sc.C.Hosts[target]; ok {
		return &HostConfig{
			Username:    hostConfig.Username,
			Password:    hostConfig.Password,
			Collectlogs: hostConfig.CollectLogs(),
			Logcount:    hostConfig.LogCount(),
		}, nil
	}
	if hostConfig, ok := sc.C.Hosts["default"]; ok {
		return &HostConfig{
			Username:    hostConfig.Username,
			Password:    hostConfig.Password,
			Collectlogs: hostConfig.CollectLogs(),
			Logcount:    hostConfig.LogCount(),
		}, nil
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
			Collectlogs: hostConfig.CollectLogs(),
			Logcount:    hostConfig.LogCount(),
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

func (hc *HostConfigYaml) CollectLogs() bool {
	if hc.Collectlogs == nil {
		return true
	}
	return *hc.Collectlogs
}

func (hc *HostConfigYaml) LogCount() int {
	if hc.Logcount == nil {
		return 0
	}
	return *hc.Logcount
}
