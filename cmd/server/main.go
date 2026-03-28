package main

import (
	"log/slog"
	"os"

	"github.com/jije/couchdb-mcp/internal/config"
	"github.com/jije/couchdb-mcp/internal/couchdb"
	"github.com/jije/couchdb-mcp/internal/mcp"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize CouchDB client
	dbClient := couchdb.NewClient(cfg)

	// Initialize MCP server
	server := mcp.NewServer(dbClient)

	slog.Info("starting MCP server", "url", cfg.CouchDBURL)

	// Start the server
	if err := server.Serve(); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
