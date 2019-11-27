package main

import (
	"fmt"
	"github.com/prometheus/common/log"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"sync"
)

type Config struct {
	Hosts map[string]HostConfig `yaml:"hosts"`
}

type SafeConfig struct {
	sync.RWMutex
	C *Config
}

type HostConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func (sc *SafeConfig) ReloadConfig(configFile string) error {
	var c = &Config{}

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Errorf("Error reading config file: %s", err)
		return err
	}
	if err := yaml.Unmarshal(yamlFile, c); err != nil {
		log.Errorf("Error parsing config file: %s", err)
		return err
	}

	sc.Lock()
	sc.C = c
	sc.Unlock()

	log.Infoln("Loaded config file")
	return nil
}

func (sc *SafeConfig) HostConfigForTarget(target string) (*HostConfig, error) {
	sc.Lock()
	defer sc.Unlock()
	if hostConfig, ok := sc.C.Hosts[target]; ok {
		return &HostConfig{
			Username: hostConfig.Username,
			Password: hostConfig.Password,
		}, nil
	}
	if hostConfig, ok := sc.C.Hosts["default"]; ok {
		return &HostConfig{
			Username: hostConfig.Username,
			Password: hostConfig.Password,
		}, nil
	}
	return &HostConfig{}, fmt.Errorf("no credentials found for target %s", target)
}
