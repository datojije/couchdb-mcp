# CouchDB MCP Server for Obsidian (Remote/Railway)

A professional-grade Model Context Protocol (MCP) server written in Go that allows LLMs to read and write Obsidian notes stored in a CouchDB instance. This version supports **HTTP transport**, making it ideal for deployment on platforms like Railway.

## Features

- **Remote Access**: Uses HTTP transport (POST `/mcp`) to allow connections from remote LLM clients.
- **Secure**: Includes Bearer Token (API Key) authentication to protect your notes.
- **Automatic Sync**: Designed for [Obsidian Self-hosted LiveSync](https://github.com/vrtmrz/obsidian-livesync).
- **Modular Architecture**: Clean, testable Go code.

## Prerequisites

- [Go 1.21+](https://golang.org/doc/install)
- A CouchDB instance (e.g., hosted on Railway)
- A Railway account (for deployment)

## Configuration

The server is configured via environment variables:

| Variable | Description |
|----------|-------------|
| `COUCHDB_URL` | The full URL to your CouchDB database |
| `COUCHDB_USER` | Your CouchDB username |
| `COUCHDB_PASS` | Your CouchDB password |
| `PORT` | The port the server listens on (default: 8080) |
| `MCP_API_KEY` | **Highly Recommended**: A secret token used for Bearer authentication. |

## Deployment to Railway

1.  **Fork or Push** this repository to your GitHub.
2.  **Create a New Project** on Railway.
3.  **Connect your Repository**.
4.  **Add Environment Variables**: Ensure you set `COUCHDB_URL`, `COUCHDB_USER`, `COUCHDB_PASS`, and `MCP_API_KEY`.
5.  Railway will automatically detect the `Dockerfile` and deploy the service.

## Connecting to your Remote MCP Server

### 1. Direct HTTP Connection

Your MCP endpoint will be at: `https://your-railway-url.up.railway.app/mcp`

### 2. Authorization

Clients must include the following header:
```text
Authorization: Bearer <YOUR_MCP_API_KEY>
```

### 3. Example Client Config (Claude Desktop Proxy)

Since Claude Desktop primarily supports `stdio`, you can use a local "proxy" script to connect it to your remote Railway server:

```json
{
  "mcpServers": {
    "remote-obsidian": {
      "command": "curl",
      "args": [
        "-X", "POST",
        "-H", "Content-Type: application/json",
        "-H", "Authorization: Bearer <YOUR_MCP_API_KEY>",
        "-d", "{\"jsonrpc\":\"2.0\",\"method\":\"notifications/initialized\",\"params\":{}}",
        "https://your-railway-url.up.railway.app/mcp"
      ]
    }
  }
}
```
*(Note: A more robust proxy utility is recommended for full bidirectional support.)*

## Development

```bash
# Install dependencies
go mod tidy

# Run locally
export COUCHDB_URL="..."
export MCP_API_KEY="supersecret"
go run cmd/server/main.go
```

## License

MIT
