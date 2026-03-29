package models

import "time"

// CouchDBDoc represents a standard document in CouchDB for Obsidian LiveSync
type CouchDBDoc struct {
	ID       string `json:"_id,omitempty"`
	Rev      string `json:"_rev,omitempty"`
	Data     string `json:"data"`     // Base64 encoded content
	Path     string `json:"path"`     // Vault path (e.g., "Folder/Note.md")
	Datatype string `json:"datatype"` // e.g., "plain" or "binary"
	Mtime    int64  `json:"mtime"`    // Last modified time in ms
}

// AllDocsResponse represents the response from CouchDB's _all_docs endpoint.
type AllDocsResponse struct {
	Rows []struct {
		ID  string     `json:"id"`
		Doc CouchDBDoc `json:"doc"`
	} `json:"rows"`
}

// Note represents the high-level note data used by the MCP server.
type Note struct {
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	LastUpdate time.Time `json:"last_update"`
}

// SearchResult represents a single hit in a search operation
type SearchResult struct {
	Title   string `json:"title"`
	Snippet string `json:"snippet"` // A small portion of content around the match
}
