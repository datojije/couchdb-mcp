# CouchDB MCP Server for Obsidian

A professional-grade Model Context Protocol (MCP) server written in Go that allows LLMs (like Claude or Gemini) to read and write Obsidian notes stored in a CouchDB instance. This is specifically designed to work with the [Obsidian Self-hosted LiveSync](https://github.com/vrtmrz/obsidian-livesync) plugin.

## Features

- **Get Note**: Fetches a note by title, automatically handling the Base64 decoding used by the LiveSync plugin.
- **Update Note**: Creates or updates a note, automatically managing CouchDB `_rev` logic.
- **Modular Architecture**: Clean, testable Go code following industry standards.
- **Structured Logging**: Uses `log/slog` for production-ready observability.

## Prerequisites

- [Go 1.21+](https://golang.org/doc/install)
- A CouchDB instance (e.g., hosted on [Railway](https://railway.app/))
- Obsidian with the [Self-hosted LiveSync](https://github.com/vrtmrz/obsidian-livesync) plugin configured.

## Configuration

The server is configured via environment variables:

| Variable | Description |
|----------|-------------|
| `COUCHDB_URL` | The full URL to your CouchDB database (e.g., `https://your-instance.railway.app/your_db_name`) |
| `COUCHDB_USER` | Your CouchDB username |
| `COUCHDB_PASS` | Your CouchDB password |

## Installation & Usage

### 1. Build the server

```bash
go build -o couchdb-mcp ./cmd/server
```

### 2. Run locally

```bash
export COUCHDB_URL="https://your-couchdb-url/db_name"
export COUCHDB_USER="your-user"
export COUCHDB_PASS="your-password"

./couchdb-mcp
```

### 3. Connect to Claude Desktop

Add the following to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "obsidian-notes": {
      "command": "/absolute/path/to/couchdb-mcp",
      "env": {
        "COUCHDB_URL": "https://your-couchdb-url/db_name",
        "COUCHDB_USER": "your-user",
        "COUCHDB_PASS": "your-password"
      }
    }
  }
}
```

## Project Structure

- `cmd/server/`: Application entry point.
- `internal/config/`: Configuration management.
- `internal/couchdb/`: CouchDB client implementation.
- `internal/mcp/`: MCP server and tool definitions.
- `internal/models/`: Data models.

## License

MIT
