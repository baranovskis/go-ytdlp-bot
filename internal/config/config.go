package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
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

type Cache struct {
	TTL string `yaml:"ttl"`
}

// GetTTL returns the cache TTL duration, defaulting to 5 minutes.
func (c *Cache) GetTTL() time.Duration {
	if c.TTL == "" {
		return 5 * time.Minute
	}
	d, err := time.ParseDuration(c.TTL)
	if err != nil {
		return 5 * time.Minute
	}
	return d
}

type Database struct {
	Path string `yaml:"path"`
}

type Dashboard struct {
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type AccessControl struct {
	Enabled      bool `yaml:"enabled"`
	DefaultAllow bool `yaml:"defaultAllow"`
}

type Config struct {
	Verbose       bool          `yaml:"verbose"`
	Storage       Storage       `yaml:"storage"`
	Bot           Bot           `yaml:"bot"`
	Cache         Cache         `yaml:"cache"`
	Database      Database      `yaml:"database"`
	Dashboard     Dashboard     `yaml:"dashboard"`
	AccessControl AccessControl `yaml:"accessControl"`
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
