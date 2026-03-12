# Webhooks

WhatsGo supports two types of webhooks:

- **Per-instance** - Configured via API for each instance individually
- **Global** - Configured in `whatsgo.yaml`, receives events from all instances

## Configuration

### Per-instance webhook

```bash
curl -X POST http://localhost:8550/api/v1/instance/my-instance/webhook \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://my-server.com/webhook",
    "events": ["message.received", "connection.update"],
    "headers": { "X-Secret": "my-secret" },
    "enabled": true
  }'
```

An empty `events` array means all events will be dispatched.

### Global webhook

In `whatsgo.yaml`:

```yaml
webhook:
  global_url: "https://my-server.com/global-webhook"
  global_events: ["message.received"]
  global_headers:
    X-Secret: "my-secret"
```

## Payload format

All webhook events are delivered as `POST` requests with the following JSON body:

```json
{
  "event": "message.received",
  "instance": "my-instance",
  "data": { ... },
  "timestamp": "2026-03-12T15:04:05Z",
  "server_url": "http://localhost:8550"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `event` | string | Event type identifier |
| `instance` | string | Name of the instance that generated the event |
| `data` | object | Event-specific payload |
| `timestamp` | string | ISO 8601 timestamp |
| `server_url` | string | Base URL of this WhatsGo server |

Headers sent with every webhook request:

- `Content-Type: application/json`
- `User-Agent: WhatsGo-API/1.0`
- Any custom headers configured for the webhook

## Event types

| Event | Description |
|-------|-------------|
| `connection.update` | Connection state changed (open/close/connecting) |
| `message.received` | New message received |
| `message.sent` | Message sent by this device |
| `message.updated` | Message status updated (delivered/read) |
| `message.deleted` | Message deleted |
| `message.reaction` | Reaction received |
| `presence.update` | User online/offline/typing status changed |
| `group.update` | Group metadata changed (name, description, photo) |
| `group.participants` | Member joined/left/promoted/demoted |
| `call.received` | Incoming call received |
| `qrcode.updated` | New QR code generated for authentication |

## Retry behavior

Failed webhook deliveries are retried with exponential backoff:

| Attempt | Delay |
|---------|-------|
| 1 | 1 second |
| 2 | 2 seconds |
| 3 | 4 seconds |
| 4 | 8 seconds |
| 5 | 16 seconds |

After 5 failed attempts, the event is dropped and logged.

**Retry rules:**
- Network errors are always retried
- HTTP `5xx`, `408`, and `429` responses are retried
- HTTP `4xx` responses (except 408/429) are **not** retried (client error)
- HTTP `2xx` responses are considered successful

Up to 50 webhook deliveries can run concurrently.
