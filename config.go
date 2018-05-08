package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bass3m/middleman/docker"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Middleman struct {
		Algorithm string `yaml:"algorithm"`
	}
	Resources struct {
		Docker struct {
			Enabled bool   `yaml:"enabled"`
			Label   string `yaml:"label"`
			Network string `yaml:"network"`
		}
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

func (c Config) GetResourceUris() ([]string, error) {
	if c.Resources.Docker.Enabled == true {
		// get resources from docker
		log.Infof("Getting resources from docker")
		uris, err := docker.GetResourceUris(c.Resources.Docker.Label, c.Resources.Docker.Network)
		if err != nil {
			return []string{""}, err
		}
		return uris, nil
	} else {
		return c.Resources.Uris, nil
	}
}
