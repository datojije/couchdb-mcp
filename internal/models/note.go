package models

// CouchDBDoc represents a standard document in CouchDB.
type CouchDBDoc struct {
	ID       string `json:"_id,omitempty"`
	Rev      string `json:"_rev,omitempty"`
	Data     string `json:"data"`     // Base64 encoded content
	Path     string `json:"path"`     // Vault path
	Datatype string `json:"datatype"` // e.g., "plain" or "binary"
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
	Title   string
	Content string
}
