package main

import (
	"github.com/prometheus/common/log"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
)

type RedfishHost struct {
	Host     string `yaml:"host"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func loadFilerFromFile(fileName string) (c []*RedfishHost) {
	var fb []RedfishHost
	yamlFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal("[ERROR] ", err)
	}
	err = yaml.Unmarshal(yamlFile, &fb)
	if err != nil {
		log.Fatal("[ERROR] ", err)
	}
	for _, b := range fb {
		c = append(c, &b)
	}
	return
}
