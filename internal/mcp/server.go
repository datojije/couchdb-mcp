package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jije/couchdb-mcp/internal/couchdb"
	"github.com/jije/couchdb-mcp/internal/models"
	"github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/http"
)

// Server encapsulates the MCP server and its dependencies.
type Server struct {
	mcpServer *mcp_golang.Server
	dbClient  couchdb.Client
	transport *http.GinTransport
}

// GetNoteArgs defines the arguments for the get_note tool.
type GetNoteArgs struct {
	Title string `json:"title" jsonschema:"required,description=The title or path of the note to fetch (e.g. 'MyNote.md')"`
}

// UpdateNoteArgs defines the arguments for the update_note tool.
type UpdateNoteArgs struct {
	Title   string `json:"title" jsonschema:"required,description=The title or path of the note to update (e.g. 'MyNote.md')"`
	Content string `json:"content" jsonschema:"required,description=The markdown content of the note"`
}

// ListNotesArgs defines the arguments for the list_notes tool.
type ListNotesArgs struct {
}

// PingArgs defines the arguments for the ping_couchdb tool.
type PingArgs struct {
}

// NewServer initializes and configures the MCP server with HTTP/Gin transport.
func NewServer(dbClient couchdb.Client) *Server {
	transport := http.NewGinTransport()
	mcpServer := mcp_golang.NewServer(transport)
	s := &Server{
		mcpServer: mcpServer,
		dbClient:  dbClient,
		transport: transport,
	}
	s.registerTools()
	return s
}

// Handler returns the Gin handler for the MCP server.
func (s *Server) Handler() gin.HandlerFunc {
	return s.transport.Handler()
}

func (s *Server) registerTools() {
	s.mcpServer.RegisterTool("get_note", "Fetches a note from CouchDB", func(args GetNoteArgs) (*mcp_golang.ToolResponse, error) {
		note, err := s.dbClient.GetNote(context.Background(), args.Title)
		if err != nil {
			return nil, fmt.Errorf("getting note: %w", err)
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(note.Content)), nil
	})

	s.mcpServer.RegisterTool("update_note", "Updates or creates a note in CouchDB", func(args UpdateNoteArgs) (*mcp_golang.ToolResponse, error) {
		err := s.dbClient.UpdateNote(context.Background(), &models.Note{
			Title:   args.Title,
			Content: args.Content,
		})
		if err != nil {
			return nil, fmt.Errorf("updating note: %w", err)
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent("Note updated successfully")), nil
	})

	s.mcpServer.RegisterTool("list_notes", "Lists all note titles/paths in CouchDB", func(args ListNotesArgs) (*mcp_golang.ToolResponse, error) {
		notes, err := s.dbClient.ListNotes(context.Background())
		if err != nil {
			return nil, fmt.Errorf("listing notes: %w", err)
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent("Available notes:\n" + strings.Join(notes, "\n"))), nil
	})

	s.mcpServer.RegisterTool("ping_couchdb", "Checks connection to CouchDB and returns the target URL", func(args PingArgs) (*mcp_golang.ToolResponse, error) {
		reqURL, err := s.dbClient.Ping(context.Background())
		if err != nil {
			return nil, fmt.Errorf("ping failed for URL %s: %w", reqURL, err)
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Successfully connected to CouchDB at %s", reqURL))), nil
	})
}

// Serve starts the internal MCP server state.
func (s *Server) Serve() error {
	return s.mcpServer.Serve()
}
