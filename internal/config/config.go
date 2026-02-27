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

type Video struct {
	MaxHeight int    `yaml:"maxHeight"`
	Threads   int    `yaml:"threads"`
	Encoder   string `yaml:"encoder"`
}

// GetMaxHeight returns the max video height, defaulting to 720.
func (v *Video) GetMaxHeight() int {
	if v.MaxHeight <= 0 {
		return 720
	}
	return v.MaxHeight
}

// GetThreads returns the ffmpeg thread count, defaulting to 2.
func (v *Video) GetThreads() int {
	if v.Threads <= 0 {
		return 2
	}
	return v.Threads
}

// GetEncoder returns the video encoder, defaulting to "libx264" (CPU).
// Supported: libx264, h264_nvenc, h264_vaapi, h264_qsv.
func (v *Video) GetEncoder() string {
	switch v.Encoder {
	case "h264_nvenc", "h264_vaapi", "h264_qsv":
		return v.Encoder
	default:
		return "libx264"
	}
}

type Config struct {
	Verbose   bool      `yaml:"verbose"`
	Storage   Storage   `yaml:"storage"`
	Bot       Bot       `yaml:"bot"`
	Cache     Cache     `yaml:"cache"`
	Database  Database  `yaml:"database"`
	Dashboard Dashboard `yaml:"dashboard"`
	Video     Video     `yaml:"video"`
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
