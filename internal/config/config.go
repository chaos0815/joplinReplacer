package config

import (
	"fmt"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	// API settings
	Host    string
	Port    int
	Token   string
	Timeout time.Duration

	// Search and replace settings
	SearchPattern string
	Replacement   string
	IsRegex       bool
	CaseSensitive bool

	// Filter settings
	NotebookID string

	// Operation settings
	DryRun      bool
	Verbose     bool
	Concurrency int
	Delay       time.Duration
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Token == "" {
		return fmt.Errorf("API token is required. Provide via --token flag or JOPLIN_TOKEN environment variable")
	}

	if c.SearchPattern == "" {
		return fmt.Errorf("search pattern is required")
	}

	if c.Host == "" {
		return fmt.Errorf("host is required")
	}

	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	if c.Concurrency < 1 || c.Concurrency > 20 {
		return fmt.Errorf("concurrency must be between 1 and 20")
	}

	if c.Delay < 0 {
		return fmt.Errorf("delay must be non-negative")
	}

	return nil
}

// NewConfig creates a new configuration with defaults
func NewConfig() *Config {
	return &Config{
		Host:          "localhost",
		Port:          41184,
		Timeout:       30 * time.Second,
		CaseSensitive: false,
		IsRegex:       false,
		DryRun:        false,
		Verbose:       false,
		Concurrency:   5,
		Delay:         100 * time.Millisecond,
	}
}
