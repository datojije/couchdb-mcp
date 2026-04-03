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
	Port        string
	MCPAPIKey   string
	PublicURL   string // Required for SSE transport (e.g. https://you.up.railway.app)
}

// LoadConfig fetches and validates configuration from environment variables.
func LoadConfig() (*Config, error) {
	url := os.Getenv("COUCHDB_URL")
	if url == "" {
		return nil, errors.New("COUCHDB_URL is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	publicURL := os.Getenv("PUBLIC_URL")
	if publicURL == "" {
		publicURL = "http://localhost:" + port
	}

	// AAA Grade: Ensure PublicURL has a scheme
	if !strings.HasPrefix(publicURL, "http://") && !strings.HasPrefix(publicURL, "https://") {
		if strings.HasPrefix(publicURL, "localhost") || strings.HasPrefix(publicURL, "127.0.0.1") {
			publicURL = "http://" + publicURL
		} else {
			publicURL = "https://" + publicURL
		}
	}

	return &Config{
		CouchDBURL:  strings.TrimSuffix(url, "/"),
		CouchDBUser: os.Getenv("COUCHDB_USER"),
		CouchDBPass: os.Getenv("COUCHDB_PASS"),
		Port:        port,
		MCPAPIKey:   os.Getenv("MCP_API_KEY"),
		PublicURL:   strings.TrimSuffix(publicURL, "/"),
	}, nil
}
