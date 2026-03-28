package models

// CouchDBDoc represents a standard document in CouchDB.
type CouchDBDoc struct {
	ID   string `json:"_id,omitempty"`
	Rev  string `json:"_rev,omitempty"`
	Data string `json:"data"` // Base64 encoded content
}

// Note represents the high-level note data used by the MCP server.
type Note struct {
	Title   string
	Content string
}
