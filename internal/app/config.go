package app

import (
	"fmt"
	"log"
	"os"
	"os/user"

	"gopkg.in/yaml.v2"
)

const (
	CONFIG_PATH = "/.config/kade/config.yml"
)

type DatabaseConfig struct {
	Host string `yaml:"host"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

type ContainerRegistryConfig struct {
	Uri  string `yaml:"uri"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

type GlobalConfig struct {
	ContainerRegistry ContainerRegistryConfig `yaml:"containerRegistry"`
	Database          DatabaseConfig          `yaml:"database"`
}

type Config struct {
	Global GlobalConfig `yaml:"global"`
}

func GetConfig() *Config {
	cfg, err := readConfig(getUserHomeDir() + CONFIG_PATH)

	if err != nil {
		fmt.Print("Could not load app config, will ask for required information...\n\n")
		return &Config{}
	}

	fmt.Print("Loaded app config from ~/.config/kade/config.yml\n\n")
	return cfg
}

func readConfig(filename string) (*Config, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func getUserHomeDir() string {
	currentUser, err := user.Current()

	if err != nil {
		log.Fatal("Could not find current users home directory")
	}

	return currentUser.HomeDir
}
