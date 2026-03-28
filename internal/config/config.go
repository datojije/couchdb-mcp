package config

import (
	"errors"
	"os"
	"strings"
)

// Config holds the application configuration.
type Config struct {
	CouchDBURL  string
	CouchDBUser string
	CouchDBPass string
}

// LoadConfig fetches and validates configuration from environment variables.
func LoadConfig() (*Config, error) {
	url := os.Getenv("COUCHDB_URL")
	if url == "" {
		return nil, errors.New("COUCHDB_URL is required")
	}

	return &Config{
		CouchDBURL:  strings.TrimSuffix(url, "/"),
		CouchDBUser: os.Getenv("COUCHDB_USER"),
		CouchDBPass: os.Getenv("COUCHDB_PASS"),
	}, nil
}
