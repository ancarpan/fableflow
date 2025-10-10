package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// Config represents the application configuration
type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"server"`
	Library struct {
		ScanDirectory string `yaml:"scan_directory"`
		AutoScan      bool   `yaml:"auto_scan"`
	} `yaml:"library"`
	Database struct {
		Path string `yaml:"path"`
	} `yaml:"database"`
}

// LoadConfig loads configuration from YAML file
func LoadConfig(filename string) (*Config, error) {
	// Set defaults
	config := &Config{}
	config.Server.Host = "localhost"
	config.Server.Port = "8080"
	config.Library.ScanDirectory = "/home/user/Books"
	config.Library.AutoScan = false
	config.Database.Path = "./ebooks.db"

	// Check if config file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Printf("Config file %s not found, using defaults", filename)
		return config, nil
	}

	// Read and parse config file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	log.Printf("Loaded configuration from %s", filename)
	return config, nil
}
