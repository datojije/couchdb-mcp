package couchdb

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/jije/couchdb-mcp/internal/config"
	"github.com/jije/couchdb-mcp/internal/models"
)

// Client handles communication with CouchDB.
type Client interface {
	GetNote(ctx context.Context, title string) (*models.Note, error)
	UpdateNote(ctx context.Context, note *models.Note) error
	ListNotes(ctx context.Context) ([]string, error)
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

func (c *client) Ping(ctx context.Context) (string, error) {
	reqURL, err := url.JoinPath(c.config.CouchDBURL, "/")
	if err != nil {
		return c.config.CouchDBURL, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return reqURL, err
	}

	if c.config.CouchDBUser != "" {
		req.SetBasicAuth(c.config.CouchDBUser, c.config.CouchDBPass)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return reqURL, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return reqURL, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	return reqURL, nil
}

func (c *client) ListNotes(ctx context.Context) ([]string, error) {
	u, err := url.Parse(c.config.CouchDBURL)
	if err != nil {
		return nil, err
	}
	u.Path = strings.TrimSuffix(u.Path, "/") + "/_all_docs"
	q := u.Query()
	q.Set("include_docs", "true")
	u.RawQuery = q.Encode()
	reqURL := u.String()

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
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
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var result models.AllDocsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var titles []string
	for _, row := range result.Rows {
		// Filter for documents that have a path (parent docs)
		// We remove the strict Datatype check to be more compatible
		if row.Doc.Path != "" && !strings.HasPrefix(row.ID, "_") {
			titles = append(titles, row.Doc.Path)
		}
	}

	return titles, nil
}

func (c *client) GetNote(ctx context.Context, title string) (*models.Note, error) {
	doc, err := c.fetchDoc(ctx, title)
	if err != nil {
		return nil, fmt.Errorf("fetching doc: %w", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(doc.Data)
	if err != nil {
		return nil, fmt.Errorf("decoding base64 data: %w", err)
	}

	return &models.Note{
		Title:   title,
		Content: string(decoded),
	}, nil
}

func (c *client) UpdateNote(ctx context.Context, note *models.Note) error {
	existingDoc, _ := c.fetchDoc(ctx, note.Title)

	newDoc := models.CouchDBDoc{
		Data: base64.StdEncoding.EncodeToString([]byte(note.Content)),
		Path: note.Title, // Set the path for the new document
	}
	if existingDoc != nil {
		newDoc.Rev = existingDoc.Rev
		newDoc.Datatype = existingDoc.Datatype
	} else {
		newDoc.Datatype = "plain"
	}

	return c.putDoc(ctx, note.Title, newDoc)
}

func (c *client) fetchDoc(ctx context.Context, id string) (*models.CouchDBDoc, error) {
	reqURL, err := url.JoinPath(c.config.CouchDBURL, url.PathEscape(id))
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
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

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var doc models.CouchDBDoc
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, err
	}

	return &doc, nil
}

func (c *client) putDoc(ctx context.Context, id string, doc models.CouchDBDoc) error {
	reqURL, err := url.JoinPath(c.config.CouchDBURL, url.PathEscape(id))
	if err != nil {
		return err
	}
	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", reqURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.config.CouchDBUser != "" {
		req.SetBasicAuth(c.config.CouchDBUser, c.config.CouchDBPass)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
