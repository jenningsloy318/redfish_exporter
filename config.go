package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"sync"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Hosts map[string]HostConfig `yaml:"hosts"`
}

type SafeConfig struct {
	sync.RWMutex
	Config *Config
}

type HostConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
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

func (sc *SafeConfig) HostConfigForTarget(target string) (*HostConfig, error) {
	sc.Lock()
	defer sc.Unlock()

	for host, account := range sc.Config.Hosts {
		match, _ := regexp.MatchString(host, target)
		if match {
			return &HostConfig{
				Username: account.Username,
				Password: account.Password,
			}, nil
		}
	}
	// if no match found, fallback to default
	if hostConfig, ok := sc.Config.Hosts["default"]; ok {
		return &HostConfig{
			Username: hostConfig.Username,
			Password: hostConfig.Password,
		}, nil
	}

	// if no default specified, exit with an error
	return &HostConfig{}, fmt.Errorf("no credentials found for target %s", target)
}
