package main

import (
	log "github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Mush struct {
		Algorithm string `yaml:"algorithm"`
	}
	Resources struct {
		Uris []string `yaml:",flow"`
	}
}

func (c *Config) ReadConfig(configPath string) *Config {

	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Errorf("Failed to read YAML config file err:  %v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	return c
}
