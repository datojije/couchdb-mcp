package mcp

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/jije/couchdb-mcp/internal/couchdb"
	"github.com/jije/couchdb-mcp/internal/models"
)

// Server encapsulates the MCP server and its SSE transport.
type Server struct {
	mcpServer *server.MCPServer
	sseServer *server.SSEServer
	dbClient  couchdb.Client
}

// NewServer initializes and configures the MCP server with SSE transport.
func NewServer(dbClient couchdb.Client, publicURL string) *Server {
	mcpServer := server.NewMCPServer("CouchDB Obsidian Server", "1.0.0")
	
	// Use WithBaseURL option for the SSE server
	sseServer := server.NewSSEServer(mcpServer, server.WithBaseURL(publicURL))

	s := &Server{
		mcpServer: mcpServer,
		sseServer: sseServer,
		dbClient:  dbClient,
	}
	s.registerTools()
	return s
}

// HandleSSE returns the handler for the SSE stream.
func (s *Server) HandleSSE() http.Handler {
	return s.sseServer.SSEHandler()
}

// HandleMessage returns the handler for incoming JSON-RPC messages.
func (s *Server) HandleMessage() http.Handler {
	return s.sseServer.MessageHandler()
}

func (s *Server) registerTools() {
	// get_note tool
	getNoteTool := mcp.NewTool("get_note",
		mcp.WithDescription("Fetches a note from CouchDB"),
		mcp.WithString("title", mcp.Required(), mcp.Description("The title or path of the note to fetch (e.g. 'MyNote.md')")),
	)
	s.mcpServer.AddTool(getNoteTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		title := req.GetString("title", "")
		note, err := s.dbClient.GetNote(ctx, title)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("getting note: %v", err)), nil
		}
		return mcp.NewToolResultText(note.Content), nil
	})

	// update_note tool
	updateNoteTool := mcp.NewTool("update_note",
		mcp.WithDescription("Updates or creates a note in CouchDB"),
		mcp.WithString("title", mcp.Required(), mcp.Description("The title or path of the note to update (e.g. 'MyNote.md')")),
		mcp.WithString("content", mcp.Required(), mcp.Description("The markdown content of the note")),
	)
	s.mcpServer.AddTool(updateNoteTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		title := req.GetString("title", "")
		content := req.GetString("content", "")
		err := s.dbClient.UpdateNote(ctx, &models.Note{
			Title:   title,
			Content: content,
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("updating note: %v", err)), nil
		}
		return mcp.NewToolResultText("Note updated successfully"), nil
	})

	// list_notes tool
	listNotesTool := mcp.NewTool("list_notes",
		mcp.WithDescription("Lists all note titles/paths in CouchDB"),
	)
	s.mcpServer.AddTool(listNotesTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		notes, err := s.dbClient.ListNotes(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("listing notes: %v", err)), nil
		}
		return mcp.NewToolResultText("Available notes:\n" + strings.Join(notes, "\n")), nil
	})

	// ping_couchdb tool
	pingTool := mcp.NewTool("ping_couchdb",
		mcp.WithDescription("Checks connection to CouchDB and returns the target URL"),
	)
	s.mcpServer.AddTool(pingTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		reqURL, err := s.dbClient.Ping(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("ping failed for URL %s: %v", reqURL, err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Successfully connected to CouchDB at %s", reqURL)), nil
	})

	// search_notes tool
	searchTool := mcp.NewTool("search_notes",
		mcp.WithDescription("Searches for a keyword in all notes and returns snippets"),
		mcp.WithString("query", mcp.Required(), mcp.Description("The search query to look for in note titles and content")),
	)
	s.mcpServer.AddTool(searchTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query := req.GetString("query", "")
		results, err := s.dbClient.SearchNotes(ctx, query)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
		}

		if len(results) == 0 {
			return mcp.NewToolResultText("No results found for: " + query), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Found %d results for '%s':\n\n", len(results), query))
		for _, res := range results {
			sb.WriteString(fmt.Sprintf("### %s\n%s\n\n", res.Title, res.Snippet))
		}
		return mcp.NewToolResultText(sb.String()), nil
	})
}
