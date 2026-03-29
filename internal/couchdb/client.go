package couchdb

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jije/couchdb-mcp/internal/config"
	"github.com/jije/couchdb-mcp/internal/models"
)

// Client handles communication with CouchDB.
type Client interface {
	GetNote(ctx context.Context, title string) (*models.Note, error)
	UpdateNote(ctx context.Context, note *models.Note) error
	ListNotes(ctx context.Context) ([]string, error)
	SearchNotes(ctx context.Context, query string) ([]models.SearchResult, error)
	Ping(ctx context.Context) (string, error)
}

type client struct {
	httpClient *http.Client
	config     *config.Config
}

// NewClient initializes a new CouchDB client.
func NewClient(cfg *config.Config) Client {
	return &client{
		httpClient: http.DefaultClient,
		config:     cfg,
	}
}

func (c *client) SearchNotes(ctx context.Context, query string) ([]models.SearchResult, error) {
	// 1. Get all documents with their full content
	u, _ := url.Parse(c.config.CouchDBURL)
	u.Path = strings.TrimSuffix(u.Path, "/") + "/_all_docs"
	q := u.Query()
	q.Set("include_docs", "true")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.config.CouchDBUser != "" {
		req.SetBasicAuth(c.config.CouchDBUser, c.config.CouchDBPass)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("couchdb status %d", resp.StatusCode)
	}

	var all models.AllDocsResponse
	if err := json.NewDecoder(resp.Body).Decode(&all); err != nil {
		return nil, err
	}

	// 2. Filter and search in-memory
	var results []models.SearchResult
	queryLower := strings.ToLower(query)

	for _, row := range all.Rows {
		// Only process documents that have a path and are NOT internal CouchDB docs
		if row.Doc.Path == "" || strings.HasPrefix(row.ID, "_") {
			continue
		}

		decoded, err := base64.StdEncoding.DecodeString(row.Doc.Data)
		if err != nil {
			continue
		}

		content := string(decoded)
		contentLower := strings.ToLower(content)

		if strings.Contains(contentLower, queryLower) || strings.Contains(strings.ToLower(row.Doc.Path), queryLower) {
			// Find snippet
			snippet := ""
			idx := strings.Index(contentLower, queryLower)
			if idx != -1 {
				start := idx - 50
				if start < 0 {
					start = 0
				}
				end := idx + len(queryLower) + 50
				if end > len(content) {
					end = len(content)
				}
				snippet = "..." + content[start:end] + "..."
			} else {
				// Match was in title/path
				if len(content) > 100 {
					snippet = content[:100] + "..."
				} else {
					snippet = content
				}
			}

			results = append(results, models.SearchResult{
				Title:   row.Doc.Path,
				Snippet: snippet,
			})
		}
	}

	return results, nil
}

func (c *client) ListNotes(ctx context.Context) ([]string, error) {
	u, _ := url.Parse(c.config.CouchDBURL)
	u.Path = strings.TrimSuffix(u.Path, "/") + "/_all_docs"
	q := u.Query()
	q.Set("include_docs", "true")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	if c.config.CouchDBUser != "" {
		req.SetBasicAuth(c.config.CouchDBUser, c.config.CouchDBPass)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result models.AllDocsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var titles []string
	for _, row := range result.Rows {
		if row.Doc.Path != "" && !strings.HasPrefix(row.ID, "_") {
			titles = append(titles, row.Doc.Path)
		}
	}
	return titles, nil
}

func (c *client) Ping(ctx context.Context) (string, error) {
	reqURL, _ := url.JoinPath(c.config.CouchDBURL, "/")
	req, _ := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if c.config.CouchDBUser != "" {
		req.SetBasicAuth(c.config.CouchDBUser, c.config.CouchDBPass)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return reqURL, err
	}
	defer resp.Body.Close()
	return reqURL, nil
}

func (c *client) GetNote(ctx context.Context, title string) (*models.Note, error) {
	u, _ := url.Parse(c.config.CouchDBURL)
	u.Path = strings.TrimSuffix(u.Path, "/") + "/_all_docs"
	q := u.Query()
	q.Set("include_docs", "true")
	u.RawQuery = q.Encode()

	req, _ := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if c.config.CouchDBUser != "" {
		req.SetBasicAuth(c.config.CouchDBUser, c.config.CouchDBPass)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var all models.AllDocsResponse
	if err := json.NewDecoder(resp.Body).Decode(&all); err != nil {
		return nil, err
	}

	for _, row := range all.Rows {
		if row.Doc.Path == title {
			decoded, _ := base64.StdEncoding.DecodeString(row.Doc.Data)
			return &models.Note{
				Title:      row.Doc.Path,
				Content:    string(decoded),
				LastUpdate: time.Unix(row.Doc.Mtime/1000, 0),
			}, nil
		}
	}

	return nil, fmt.Errorf("note not found: %s", title)
}

func (c *client) UpdateNote(ctx context.Context, note *models.Note) error {
	u, _ := url.Parse(c.config.CouchDBURL)
	u.Path = strings.TrimSuffix(u.Path, "/") + "/_all_docs"
	q := u.Query()
	q.Set("include_docs", "true")
	u.RawQuery = q.Encode()

	req, _ := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if c.config.CouchDBUser != "" {
		req.SetBasicAuth(c.config.CouchDBUser, c.config.CouchDBPass)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var all models.AllDocsResponse
	json.NewDecoder(resp.Body).Decode(&all)

	var existingDoc *models.CouchDBDoc
	for _, row := range all.Rows {
		if row.Doc.Path == note.Title {
			existingDoc = &row.Doc
			break
		}
	}

	newDoc := models.CouchDBDoc{
		Data:     base64.StdEncoding.EncodeToString([]byte(note.Content)),
		Path:     note.Title,
		Datatype: "plain",
		Mtime:    time.Now().UnixMilli(),
	}

	docID := ""
	if existingDoc != nil {
		newDoc.Rev = existingDoc.Rev
		docID = existingDoc.ID
		newDoc.Datatype = existingDoc.Datatype
	} else {
		docID = url.PathEscape(note.Title)
	}

	return c.putDoc(ctx, docID, newDoc)
}

func (c *client) putDoc(ctx context.Context, id string, doc models.CouchDBDoc) error {
	reqURL, _ := url.JoinPath(c.config.CouchDBURL, id)
	body, _ := json.Marshal(doc)
	req, _ := http.NewRequestWithContext(ctx, "PUT", reqURL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	if c.config.CouchDBUser != "" {
		req.SetBasicAuth(c.config.CouchDBUser, c.config.CouchDBPass)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
