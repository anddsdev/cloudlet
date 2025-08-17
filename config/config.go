package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port    string `yaml:"port"`
		Storage struct {
			Path string `yaml:"path"`
		} `yaml:"storage"`
		Timeout struct {
			ReadTimeout       int `yaml:"read_timeout"`
			WriteTimeout      int `yaml:"write_timeout"`
			IdleTimeout       int `yaml:"idle_timeout"`
			MaxHeaderBytes    int `yaml:"max_header_bytes"`
			ReadHeaderTimeout int `yaml:"read_header_timeout"`
		} `yaml:"timeout"`
	} `yaml:"server"`

	Database struct {
		DSN     string `yaml:"dsn"`
		MaxConn int    `yaml:"max_conn"`
	} `yaml:"database"`
}

func NewConfig(cfgPath string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(cfgPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}
