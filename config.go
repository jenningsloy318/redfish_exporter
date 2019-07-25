package main

import (
	"fmt"
	"github.com/prometheus/common/log"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"sync"
)

type Config struct {
	Credentials map[string]Credential `yaml:"credentials"`
}

type SafeConfig struct {
	sync.RWMutex
	C *Config
}

type Credential struct {
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

func (sc *SafeConfig) CredentialsForTarget(target string) (*Credential, error) {
	sc.Lock()
	defer sc.Unlock()
	if credential, ok := sc.C.Credentials[target]; ok {
		return &Credential{
			Username: credential.Username,
			Password: credential.Password,
		}, nil
	}
	if credential, ok := sc.C.Credentials["default"]; ok {
		return &Credential{
			Username: credential.Username,
			Password: credential.Password,
		}, nil
	}
	return &Credential{}, fmt.Errorf("no credentials found for target %s", target)
}
