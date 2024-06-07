package config

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"

	"github.com/adde/kade/internal/prompts"
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

func CreateConfig() {
	if fileExists(getUserHomeDir() + CONFIG_PATH) {
		fmt.Println("Config file already exists, aborting...")
		os.Exit(0)
	}

	fmt.Print("Creating config file...\n\n")

	cfg := Config{
		Global: GlobalConfig{
			ContainerRegistry: ContainerRegistryConfig{
				Uri:  prompts.TextInput("Container registry URI?", "", "", false),
				User: prompts.TextInput("Container registry user?", "", "", false),
				Pass: prompts.PassWordInput("Container registry password?", "", "", false),
			},
			Database: DatabaseConfig{
				Host: prompts.TextInput("Database host?", "", "", false),
				User: prompts.TextInput("Database user?", "", "", false),
				Pass: prompts.PassWordInput("Database password?", "", "", false),
			},
		},
	}

	err := writeConfig(getUserHomeDir()+CONFIG_PATH, &cfg)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nConfig file created!")
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

func writeConfig(filename string, cfg *Config) error {
	dir := strings.ReplaceAll(filename, "/config.yml", "")

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)

		if err != nil {
			return err
		}
	}

	buf, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, buf, 0644)
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	// File may exist but there's an error accessing it (e.g. permissions issue)
	return false
}

func getUserHomeDir() string {
	currentUser, err := user.Current()

	if err != nil {
		log.Fatal("Could not find current users home directory")
	}

	return currentUser.HomeDir
}
