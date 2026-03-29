# CouchDB MCP Server for Obsidian (Remote/Railway)

A professional-grade Model Context Protocol (MCP) server written in Go that allows LLMs to read and write Obsidian notes stored in a CouchDB instance. This is specifically designed for use with the [Obsidian Self-hosted LiveSync](https://github.com/vrtmrz/obsidian-livesync) plugin and optimized for remote deployment on **Railway**.

## Features

- **Authenticated HTTP Transport**: Securely expose your MCP server to remote clients using Bearer tokens.
- **CouchDB Integration**: Seamlessly read and write notes to your CouchDB database.
- **Base64 Management**: Automatically handles Base64 encoding/decoding for note content, as required by the LiveSync plugin.
- **Clean Architecture**: Modular Go design (internal/config, internal/couchdb, internal/mcp) for maximum maintainability.
- **Docker Ready**: Includes a multi-stage `Dockerfile` for efficient deployments.

## Prerequisites

- [Go 1.25.7+](https://golang.org/doc/install)
- A CouchDB instance (hosted on Railway or elsewhere)
- A Railway account for hosting this MCP server

## Configuration

The server is configured entirely via environment variables:

| Variable | Requirement | Description |
|----------|-------------|-------------|
| `COUCHDB_URL` | **Required** | The full URL to your CouchDB database (e.g., `https://.../db_name`) |
| `COUCHDB_USER` | **Optional** | Your CouchDB username (if authentication is enabled) |
| `COUCHDB_PASS` | **Optional** | Your CouchDB password (if authentication is enabled) |
| `MCP_API_KEY` | **Required** | Your secret token. Requests MUST include `Authorization: Bearer <token>` |
| `PORT` | Optional | The port to listen on (default: `8080`, automatically set by Railway) |
| `GIN_MODE` | Optional | Set to `release` for production (default: `release`) |

## Deployment to Railway

1. **GitHub Connection**: Push this repository to your GitHub account.
2. **New Service**: On Railway, click **New** -> **GitHub Repo** and select this repository.
3. **Variables**: Go to the **Variables** tab and add all the required environment variables mentioned above.
4. **Deploy**: Railway will detect the `Dockerfile` and deploy the server automatically.

## How to Connect

Once deployed, your server endpoint is: `https://your-service-name.up.railway.app/mcp`

### Security (Authentication)

Every request to your MCP server must include an `Authorization` header. If the header is missing or the token doesn't match your `MCP_API_KEY`, the server will return `401 Unauthorized`.

**Header Format:**
```text
Authorization: Bearer YOUR_MCP_API_KEY
```

## Local Development

```bash
# Clone the repository
git clone git@github.com:datojije/couchdb-mcp.git
cd couchdb-mcp

# Install dependencies
go mod tidy

# Set environment variables and run
export COUCHDB_URL="https://..."
export MCP_API_KEY="my-secret-key"
go run cmd/server/main.go
```

## License

MIT
