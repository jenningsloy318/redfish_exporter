package config

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"sync"

	yaml "gopkg.in/yaml.v2"
)

//Config is the struct that represents config.yml
type Config struct {
	Timeout  int32                 `yaml:"timeout"`
	Insecure bool                  `yaml:"insecure"`
	Port     int32                 `yaml:"port"`
	Hosts    map[string]HostConfig `yaml:"hosts"`
}

type SafeConfig struct {
	sync.RWMutex
	Config *Config
}

//HostConfig is the struct that is passed to the redfish client
type HostConfig struct {
	Host     string
	Username string
	Password string
	Timeout  int32
	Insecure bool
}

func (sc *SafeConfig) ReloadConfig(configFile string) error {
	var c = &Config{}

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(yamlFile, c); err != nil {
		return err
	}

	sc.Lock()
	sc.Config = c
	sc.Unlock()

	return nil
}

func (sc *SafeConfig) HostConfigForTarget(targetHost string) (*HostConfig, error) {
	sc.Lock()
	defer sc.Unlock()

	targetHostConfig := &HostConfig{
		Host:     targetHost,
		Timeout:  sc.Config.Timeout,
		Insecure: sc.Config.Insecure,
	}

	// search for host that matches the target
	for host, account := range sc.Config.Hosts {
		match, _ := regexp.MatchString(host, targetHost)
		if match {
			targetHostConfig.Username = account.Username
			targetHostConfig.Password = account.Password
			return targetHostConfig, nil
		}
	}
	// if no match found, fallback to default
	if hostConfig, ok := sc.Config.Hosts["default"]; ok {
		targetHostConfig.Username = hostConfig.Username
		targetHostConfig.Password = hostConfig.Password
		return targetHostConfig, nil
	}

	// if no default specified, exit with an error
	return &HostConfig{}, fmt.Errorf("no credentials found for target %s", targetHost)
}
