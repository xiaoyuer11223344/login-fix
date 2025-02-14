package ocr

import (
	"errors"
	"time"
)

type Config struct {
	BaseURL     string
	MaxAttempts int
	Timeout     time.Duration
}

func DefaultConfig() *Config {
	return &Config{}
}

// ValidateConfig ensures the configuration is valid
func (c *Config) ValidateConfig() error {
	if c.BaseURL == "" {
		return errors.New("OCR base URL must be provided through environment configuration")
	}
	return nil
}

func NewConfigWithURL(baseURL string) *Config {
	cfg := DefaultConfig()
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	return cfg
}
