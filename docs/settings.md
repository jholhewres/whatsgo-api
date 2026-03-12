# Per-Instance Settings

Each instance can be configured independently via the settings API.

## Available settings

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `reject_call` | bool | `false` | Automatically reject incoming calls |
| `reject_call_message` | string | `""` | Text message sent when rejecting a call |
| `groups_ignore` | bool | `false` | Ignore all group events (messages, updates, etc.) |
| `always_online` | bool | `false` | Keep presence set to "online" at all times |
| `read_messages` | bool | `false` | Automatically mark incoming messages as read |
| `read_receipts` | bool | `false` | Send read receipts (blue checkmarks) |
| `webhook_base64` | bool | `false` | Include media content as base64 in webhook payloads |

## Get settings

```bash
curl http://localhost:8550/api/v1/instance/my-instance/settings \
  -H "Authorization: Bearer <token>"
```

Response:

```json
{
  "instance_id": "uuid",
  "reject_call": false,
  "reject_call_message": "",
  "groups_ignore": false,
  "always_online": false,
  "read_messages": false,
  "read_receipts": false,
  "webhook_base64": false
}
```

## Update settings

Only include the fields you want to change:

```bash
curl -X PUT http://localhost:8550/api/v1/instance/my-instance/settings \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "always_online": true,
    "reject_call": true,
    "reject_call_message": "Can'\''t talk right now, send me a message."
  }'
```

## Notes

- `reject_call_message` only takes effect when `reject_call` is `true`
- `webhook_base64` can significantly increase webhook payload sizes for media messages
- `groups_ignore` filters events at the application level; the WhatsApp connection still receives group data
- Settings are persisted in the database and survive restarts
