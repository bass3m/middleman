package config

import (
	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

type Config struct {
	Client     *docker.Client
	FileConfig FileConfig
}

type FileConfig struct {
	Middleman struct {
		Algorithm string `yaml:"algorithm"`
	}
	Resources struct {
		Docker struct {
			Enabled      bool          `yaml:"enabled"`
			Retries      int           `yaml:"retries"`
			RetryTimeout time.Duration `yaml:"retry_timeout"`
			Endpoint     string        `yaml:"endpoint"`
			Label        string        `yaml:"label"`
			Network      string        `yaml:"network"`
		}
		Uris []string `yaml:",flow"`
	}
}

func ReadConfig(configPath string) (Config, error) {
	var cfg Config

	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Errorf("Failed to read YAML config file err:  %v ", err)
		return Config{}, err
	}
	err = yaml.Unmarshal(yamlFile, &cfg.FileConfig)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
		return Config{}, err
	}

	return cfg, nil
}
