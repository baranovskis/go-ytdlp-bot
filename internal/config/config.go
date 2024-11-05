package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Storage struct {
	Path             string `yaml:"path"`
	RemoveAfterReply bool   `yaml:"removeAfterReply"`
}

type Bot struct {
	Token  string      `yaml:"token"`
	Filter []BotFilter `yaml:"filters"`
}

type BotFilter struct {
	ExcludeQueryParams bool     `yaml:"excludeQueryParams"`
	Hosts              []string `yaml:"hosts"`
	PathRegEx          string   `yaml:"pathRegEx"`
	CookiesFile        string   `yaml:"cookiesFile"`
}

type Config struct {
	Verbose bool    `yaml:"verbose"`
	Storage Storage `yaml:"storage"`
	Bot     Bot     `yaml:"bot"`
}

func GetConfiguration(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
