package config

import (
	"os"
)

type DocsConfig struct {
	User     string
	Password string
}

func DocsLoad() *DocsConfig {
	config := &DocsConfig{
		User:     os.Getenv("DOCS_USER"),
		Password: os.Getenv("DOCS_PASSWORD"),
	}

	if config.User == "" {
		config.User = "root"
	}

	if config.Password == "" {
		config.Password = "root"
	}

	return config
}
