# CouchDB MCP Server for Obsidian (Remote/Railway)

A professional-grade Model Context Protocol (MCP) server written in Go that allows LLMs to read and write Obsidian notes stored in a CouchDB instance. This version supports **SSE (Server-Sent Events) transport**, which is the official standard for remote MCP connections.

## Features

- **SSE Transport**: Fully compliant with the MCP spec for remote clients (Claude, Gemini CLI, etc.)
- **Two-Step Protocol**: Uses `/sse` for the stream and `/message` for JSON-RPC commands.
- **Secure**: Includes Bearer Token (API Key) authentication.
- **Search & Discovery**: Includes `search_notes` and `list_notes` for powerful note retrieval.

## Prerequisites

- [Go 1.25.7+](https://golang.org/doc/install)
- A CouchDB instance (e.g., hosted on Railway)

## Configuration

The server is configured via environment variables:

| Variable | Requirement | Description |
|----------|-------------|-------------|
| `COUCHDB_URL` | **Required** | The full URL to your CouchDB database (e.g., `http://...:5984/db_name`) |
| `COUCHDB_USER` | Optional | Your CouchDB username |
| `COUCHDB_PASS` | Optional | Your CouchDB password |
| `MCP_API_KEY` | **Required** | Secret token. Must be sent in the `Authorization: Bearer <token>` header. |
| `PUBLIC_URL` | **Required** | Your public Railway URL (e.g., `https://your-app.up.railway.app`) |
| `PORT` | Optional | The port to listen on (default: 8080) |

## Deployment to Railway

1. **Fork or Push** this repository to your GitHub.
2. **Create a New Project** on Railway and connect your Repo.
3. **Add Variables**: Ensure you set `COUCHDB_URL`, `MCP_API_KEY`, and `PUBLIC_URL`.
4. Railway will automatically build and deploy using the `Dockerfile`.

## Connecting to Gemini CLI

To add this remote server to your Gemini CLI, use the `/sse` endpoint:

```bash
gemini mcp add --transport http --header "Authorization: Bearer <YOUR_API_KEY>" obsidian-remote https://your-app.up.railway.app/sse
```

## Available Tools

- `list_notes`: List all notes in your vault.
- `get_note(title)`: Fetch the content of a specific note.
- `update_note(title, content)`: Update or create a note.
- `search_notes(query)`: Search for a keyword in all notes.
- `ping_couchdb`: Test the connection between the MCP server and CouchDB.

## License

MIT
