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

	"github.com/jije/couchdb-mcp/internal/config"
	"github.com/jije/couchdb-mcp/internal/models"
)

// Client handles communication with CouchDB.
type Client interface {
	GetNote(ctx context.Context, title string) (*models.Note, error)
	UpdateNote(ctx context.Context, note *models.Note) error
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
	// Try to fetch existing doc for _rev
	existingDoc, _ := c.fetchDoc(ctx, note.Title)

	newDoc := models.CouchDBDoc{
		Data: base64.StdEncoding.EncodeToString([]byte(note.Content)),
	}
	if existingDoc != nil {
		newDoc.Rev = existingDoc.Rev
	}

	return c.putDoc(ctx, note.Title, newDoc)
}

func (c *client) fetchDoc(ctx context.Context, id string) (*models.CouchDBDoc, error) {
	reqURL := fmt.Sprintf("%s/%s", c.config.CouchDBURL, url.PathEscape(id))
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
	reqURL := fmt.Sprintf("%s/%s", c.config.CouchDBURL, url.PathEscape(id))
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
