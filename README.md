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

The API will be available at `http://localhost:8550`.

### Manual build

**Requirements:** Go 1.24+, Node.js 22+, GCC (for SQLite)

```bash
# Clone
git clone https://github.com/jholhewres/whatsgo-api.git
cd whatsgo-api

# Build (frontend + backend)
make build

# Run
./whatsgo
```

### Development

```bash
# Backend
make dev

# Frontend (in another terminal)
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

Environment variables are supported using the `${VAR:-default}` format.

## Authentication

Every request requires a token. Two types are supported:

| Type | Header | Scope |
|------|--------|-------|
| Global API Key | `X-API-Key: <key>` | Full access to all instances |
| Instance Token | `Authorization: Bearer <token>` | Access only to the token's instance |

The instance token is returned when creating an instance.

## Endpoints

### Instance

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/instance/create` | Create instance |
| `POST` | `/api/v1/instance/{name}/connect` | Connect (QR or pairing code) |
| `POST` | `/api/v1/instance/{name}/restart` | Restart connection |
| `GET` | `/api/v1/instance/{name}/status` | Connection status |
| `GET` | `/api/v1/instance` | List instances |
| `DELETE` | `/api/v1/instance/{name}/logout` | Logout from WhatsApp |
| `DELETE` | `/api/v1/instance/{name}` | Delete instance |

### Messages

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/instance/{name}/message/send-text` | Send text |
| `POST` | `/api/v1/instance/{name}/message/send-media` | Send media |
| `POST` | `/api/v1/instance/{name}/message/send-location` | Send location |
| `POST` | `/api/v1/instance/{name}/message/send-contact` | Send contact card |
| `POST` | `/api/v1/instance/{name}/message/send-reaction` | Send reaction |
| `POST` | `/api/v1/instance/{name}/message/send-sticker` | Send sticker |

### Chat

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/instance/{name}/chat/check-number` | Check numbers on WhatsApp |
| `POST` | `/api/v1/instance/{name}/chat/mark-read` | Mark as read |
| `POST` | `/api/v1/instance/{name}/chat/delete-message` | Delete message |
| `POST` | `/api/v1/instance/{name}/chat/edit-message` | Edit message |
| `POST` | `/api/v1/instance/{name}/chat/send-presence` | Typing/recording |
| `POST` | `/api/v1/instance/{name}/chat/block` | Block/unblock |
| `GET` | `/api/v1/instance/{name}/chat/contacts` | List contacts |
| `GET` | `/api/v1/instance/{name}/chat/profile/{jid}` | Get user profile |

### Group

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/instance/{name}/group/create` | Create group |
| `GET` | `/api/v1/instance/{name}/group` | List groups |
| `GET` | `/api/v1/instance/{name}/group/{jid}` | Get group info |
| `PUT` | `/api/v1/instance/{name}/group/{jid}/name` | Update name |
| `PUT` | `/api/v1/instance/{name}/group/{jid}/description` | Update description |
| `PUT` | `/api/v1/instance/{name}/group/{jid}/photo` | Update photo |
| `PUT` | `/api/v1/instance/{name}/group/{jid}/settings` | Update settings |
| `POST` | `/api/v1/instance/{name}/group/{jid}/participants` | Manage participants |
| `GET` | `/api/v1/instance/{name}/group/{jid}/invite-link` | Get invite link |
| `DELETE` | `/api/v1/instance/{name}/group/{jid}/leave` | Leave group |

### Webhook

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/instance/{name}/webhook` | Set webhook |
| `GET` | `/api/v1/instance/{name}/webhook` | Get webhook config |
| `DELETE` | `/api/v1/instance/{name}/webhook` | Remove webhook |

### Settings

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/instance/{name}/settings` | Get settings |
| `PUT` | `/api/v1/instance/{name}/settings` | Update settings |

## Examples

### Create instance

```bash
curl -X POST http://localhost:8550/api/v1/instance/create \
  -H "X-API-Key: your-key" \
  -H "Content-Type: application/json" \
  -d '{"name": "my-instance"}'
```

### Connect via QR Code

```bash
curl -X POST http://localhost:8550/api/v1/instance/my-instance/connect \
  -H "Authorization: Bearer <instance-token>"
```

### Connect via Pairing Code

```bash
curl -X POST http://localhost:8550/api/v1/instance/my-instance/connect \
  -H "Authorization: Bearer <instance-token>" \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+5511999999999"}'
```

### Send message

```bash
curl -X POST http://localhost:8550/api/v1/instance/my-instance/message/send-text \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"number": "5511999999999", "text": "Hello!"}'
```

### Set up webhook

```bash
curl -X POST http://localhost:8550/api/v1/instance/my-instance/webhook \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://my-server.com/webhook",
    "events": ["message.received", "connection.update"],
    "enabled": true
  }'
```

## Webhook Events

| Event | Description |
|-------|-------------|
| `connection.update` | Connection state changed (open/close/connecting) |
| `message.received` | New message received |
| `message.sent` | Message sent by this device |
| `message.updated` | Message status updated (delivered/read) |
| `message.deleted` | Message deleted |
| `message.reaction` | Reaction received |
| `presence.update` | Online/offline/typing |
| `group.update` | Group metadata changed |
| `group.participants` | Member joined/left/promoted/demoted |
| `call.received` | Incoming call |
| `qrcode.updated` | New QR code generated |

Payload format:

```json
{
  "event": "message.received",
  "instance": "my-instance",
  "data": { ... },
  "timestamp": "2026-03-12T15:04:05Z",
  "server_url": "http://localhost:8550"
}
```

## Per-Instance Settings

| Setting | Type | Description |
|---------|------|-------------|
| `reject_call` | bool | Auto-reject incoming calls |
| `reject_call_message` | string | Auto-reply message when rejecting |
| `groups_ignore` | bool | Ignore group events |
| `always_online` | bool | Keep online presence |
| `read_messages` | bool | Auto-mark messages as read |
| `read_receipts` | bool | Send read receipts |
| `webhook_base64` | bool | Include media as base64 in webhooks |

## Project Structure

```
cmd/whatsgo/          # Entry point
internal/
  api/                # HTTP handlers and types
  auth/               # Authentication middleware
  cache/              # Generic TTL cache
  config/             # YAML configuration system
  server/             # HTTP server and middlewares
  store/              # Persistence interface (PostgreSQL + SQLite)
  webhook/            # Webhook dispatcher with retry
  whatsapp/           # WhatsApp instance manager (whatsmeow)
  webui/              # Embedded React frontend
web/                  # React frontend source code
```

## Contributing

Contributions are welcome! Feel free to open an issue or pull request.

## License

This project is licensed under the [MIT License](LICENSE).
