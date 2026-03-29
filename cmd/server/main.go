package main

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jije/couchdb-mcp/internal/config"
	"github.com/jije/couchdb-mcp/internal/couchdb"
	"github.com/jije/couchdb-mcp/internal/mcp"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize CouchDB client
	dbClient := couchdb.NewClient(cfg)

	// Initialize MCP server with Gin transport
	mcpServer := mcp.NewServer(dbClient)

	// Start the internal MCP server state in a goroutine
	go func() {
		if err := mcpServer.Serve(); err != nil {
			slog.Error("MCP internal server error", "error", err)
		}
	}()

	// Set up Gin router
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// Authentication Middleware
	authMiddleware := func(c *gin.Context) {
		if cfg.MCPAPIKey == "" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be in 'Bearer <token>' format"})
			return
		}

		if parts[1] != cfg.MCPAPIKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			return
		}

		c.Next()
	}

	// Register MCP endpoint with authentication
	router.POST("/mcp", authMiddleware, mcpServer.Handler())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	slog.Info("starting HTTP MCP server", "port", cfg.Port)

	// Start the HTTP server
	if err := router.Run(":" + cfg.Port); err != nil {
		slog.Error("failed to run server", "error", err)
		os.Exit(1)
	}
}
