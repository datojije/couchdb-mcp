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

	// Initialize MCP server with SSE transport
	mcpServer := mcp.NewServer(dbClient, cfg.PublicURL)

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
			authHeader = c.Query("token")
			if authHeader != "" {
				if authHeader != cfg.MCPAPIKey {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
					return
				}
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header or token query param is required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be in 'Bearer <token>' format"})
			return
		}

		if parts[1] != cfg.MCPAPIKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
			return
		}

		c.Next()
	}

	// Delegate MCP routes to the mcpServer handler
	// AAA Grade: Allow both GET and POST for maximum compatibility
	mcpHandler := func(c *gin.Context) {
		mcpServer.ServeHTTP(c.Writer, c.Request)
	}

	router.Match([]string{"GET", "POST"}, "/sse", authMiddleware, mcpHandler)
	router.Match([]string{"GET", "POST"}, "/message", authMiddleware, mcpHandler)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	slog.Info("starting SSE MCP server", "port", cfg.Port, "public_url", cfg.PublicURL)

	// Start the HTTP server
	if err := router.Run(":" + cfg.Port); err != nil {
		slog.Error("failed to run server", "error", err)
		os.Exit(1)
	}
}
