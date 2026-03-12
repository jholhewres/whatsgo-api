# WhatsGo API

Multi-instance WhatsApp REST API built with Go and [whatsmeow](https://github.com/tulir/whatsmeow). Simple, straightforward, and production-ready.

## Features

- **Multi-instance** - Manage multiple WhatsApp connections simultaneously
- **Dual authentication** - Global API key + per-instance token
- **QR Code & Pairing Code** - Two ways to connect
- **Messages** - Text, media, location, contact, sticker, reaction
- **Groups** - Full CRUD, participants, invite links
- **Webhooks** - Per-instance + global, with exponential backoff retry
- **Per-instance settings** - Reject calls, always online, auto-read, etc.
- **Embedded frontend** - React dashboard for management
- **PostgreSQL + SQLite** - Postgres for production, SQLite for local dev

## Quick Start

### Docker (recommended)

```bash
git clone https://github.com/jholhewres/whatsgo-api.git
cd whatsgo-api
docker compose up -d
```

The API will be available at `http://localhost:8550`. Access the dashboard at the same address in your browser.

### Manual build

**Requirements:** Go 1.24+, Node.js 22+, GCC (for SQLite)

```bash
git clone https://github.com/jholhewres/whatsgo-api.git
cd whatsgo-api
make build
./whatsgo
```

### Development

```bash
# Backend
make dev

# Frontend (separate terminal)
make dev-frontend
```

## Configuration

Create a `whatsgo.yaml` file (or set `WHATSGO_CONFIG` to a different path):

```yaml
server:
  host: "0.0.0.0"
  port: 8550
  base_url: "http://localhost:8550"

database:
  backend: "postgresql"   # or "sqlite"
  postgresql:
    host: "${POSTGRES_HOST:-localhost}"
    port: 5432
    user: "${POSTGRES_USER:-postgres}"
    password: "${POSTGRES_PASSWORD:-postgres}"
    dbname: "${POSTGRES_DB:-whatsgo}"
    sslmode: "disable"
  sqlite:
    path: "./data/whatsgo.db"

auth:
  global_api_key: "${WHATSGO_API_KEY}"

whatsapp:
  session_db_path: "./data/sessions.db"

webhook:
  global_url: ""
  global_events: []
  global_headers: {}

logging:
  level: "info"
  format: "text"   # or "json"
```

Environment variables are supported using the `${VAR:-default}` format. See [`.env.example`](.env.example) for a reference.

## Authentication

Every request requires a token. Two types are supported:

| Type | Header | Scope |
|------|--------|-------|
| Global API Key | `X-API-Key: <key>` | Full access to all instances |
| Instance Token | `Authorization: Bearer <token>` | Access only to the token's instance |

The instance token is returned when creating an instance. The global API key is set in the configuration file.

## Architecture

```
cmd/whatsgo/          # Entry point
internal/
  api/                # HTTP handlers and request/response types
  auth/               # Authentication middleware (API key + instance token)
  cache/              # Generic TTL cache
  config/             # YAML configuration with env var expansion
  server/             # HTTP server, routing, and middleware chain
  store/              # Persistence layer (PostgreSQL + SQLite)
  webhook/            # Webhook dispatcher with exponential backoff
  whatsapp/           # WhatsApp instance manager (whatsmeow wrapper)
  webui/              # Embedded React frontend (compiled into binary)
web/                  # React frontend source (Vite + TypeScript)
```

The application uses two database layers:

- **WhatsGo Store** - Application tables (instances, webhooks, settings) via `pgxpool` (Postgres) or `database/sql` (SQLite)
- **whatsmeow Store** - Session/encryption tables managed internally by whatsmeow

For PostgreSQL, both share the same database. For SQLite, separate files are used.

## Documentation

Full API reference is available in the [`docs/`](docs/) directory:

- [API Reference](docs/api-reference.md) - All endpoints with request/response examples
- [Webhooks](docs/webhooks.md) - Event types, payload format, and configuration
- [Settings](docs/settings.md) - Per-instance settings reference

The embedded frontend also includes interactive API documentation at `/docs` when the server is running.

## Contributing

Contributions are welcome. Open an issue or pull request.

## License

This project is licensed under the [MIT License](LICENSE).
