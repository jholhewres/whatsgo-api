# API Reference

Base URL: `http://localhost:8550`

All endpoints are under `/api/v1/`. Authentication is required via `X-API-Key` header (global key) or `Authorization: Bearer <token>` (instance token).

---

## Instance

### Create instance

```
POST /api/v1/instance/create
```

```json
{ "name": "my-instance" }
```

Response `201`:

```json
{
  "instance": {
    "id": "uuid",
    "name": "my-instance",
    "token": "generated-token",
    "status": "created",
    "created_at": "2026-03-12T15:04:05Z"
  }
}
```

### Connect

```
POST /api/v1/instance/{name}/connect
```

**QR Code flow** (no body or empty body):

```bash
curl -X POST http://localhost:8550/api/v1/instance/my-instance/connect \
  -H "Authorization: Bearer <token>"
```

Response `200`:

```json
{ "status": "qr", "qr_code": "2@..." }
```

**Pairing Code flow** (with phone number):

```json
{ "phone_number": "+5511999999999" }
```

Response `200`:

```json
{ "status": "pairing_code", "pairing_code": "ABCD-EFGH" }
```

### Restart

```
POST /api/v1/instance/{name}/restart
```

Response `200`:

```json
{ "status": "ok", "message": "instance restarted" }
```

### Get status

```
GET /api/v1/instance/{name}/status
```

Response `200`:

```json
{
  "instance": {
    "id": "uuid",
    "name": "my-instance",
    "status": "open",
    "phone": "5511999999999",
    "business_name": "My Business",
    "created_at": "2026-03-12T15:04:05Z"
  }
}
```

### List instances

```
GET /api/v1/instance
```

Response `200`: Array of instance objects (token omitted).

### Logout

```
DELETE /api/v1/instance/{name}/logout
```

Response `200`:

```json
{ "status": "ok", "message": "logged out" }
```

### Delete instance

```
DELETE /api/v1/instance/{name}
```

Response `200`:

```json
{ "status": "ok", "message": "instance deleted" }
```

---

## Messages

All message endpoints are under `/api/v1/instance/{name}/message/`.

### Send text

```
POST /api/v1/instance/{name}/message/send-text
```

```json
{
  "number": "5511999999999",
  "text": "Hello!",
  "reply_to": "optional-message-id"
}
```

Response `200`:

```json
{
  "message_id": "3EB0...",
  "status": "sent",
  "timestamp": "2026-03-12T15:04:05Z"
}
```

### Send media

```
POST /api/v1/instance/{name}/message/send-media
```

```json
{
  "number": "5511999999999",
  "media_type": "image",
  "media": "https://example.com/photo.jpg",
  "caption": "Check this out",
  "filename": "photo.jpg",
  "mime_type": "image/jpeg"
}
```

`media_type`: `image`, `video`, `audio`, `document`

`media`: URL or base64-encoded data.

### Send location

```
POST /api/v1/instance/{name}/message/send-location
```

```json
{
  "number": "5511999999999",
  "latitude": -23.5505,
  "longitude": -46.6333,
  "name": "Sao Paulo",
  "address": "Av. Paulista, 1000"
}
```

### Send contact

```
POST /api/v1/instance/{name}/message/send-contact
```

```json
{
  "number": "5511999999999",
  "contact_name": "John Doe",
  "phones": [
    { "number": "+5511888888888", "type": "CELL" }
  ]
}
```

### Send reaction

```
POST /api/v1/instance/{name}/message/send-reaction
```

```json
{
  "number": "5511999999999",
  "message_id": "3EB0...",
  "emoji": "\ud83d\udc4d"
}
```

Send empty `emoji` to remove reaction.

### Send sticker

```
POST /api/v1/instance/{name}/message/send-sticker
```

```json
{
  "number": "5511999999999",
  "sticker": "base64-encoded-webp",
  "mime_type": "image/webp"
}
```

---

## Chat

All chat endpoints are under `/api/v1/instance/{name}/chat/`.

### Check number

```
POST /api/v1/instance/{name}/chat/check-number
```

```json
{ "numbers": ["5511999999999", "5511888888888"] }
```

Response `200`:

```json
{
  "results": [
    { "number": "5511999999999", "exists": true, "jid": "5511999999999@s.whatsapp.net" },
    { "number": "5511888888888", "exists": false }
  ]
}
```

### Mark as read

```
POST /api/v1/instance/{name}/chat/mark-read
```

```json
{
  "chat_jid": "5511999999999@s.whatsapp.net",
  "message_ids": ["3EB0..."]
}
```

### Delete message

```
POST /api/v1/instance/{name}/chat/delete-message
```

```json
{
  "chat_jid": "5511999999999@s.whatsapp.net",
  "message_id": "3EB0..."
}
```

### Edit message

```
POST /api/v1/instance/{name}/chat/edit-message
```

```json
{
  "chat_jid": "5511999999999@s.whatsapp.net",
  "message_id": "3EB0...",
  "text": "edited text"
}
```

### Send presence

```
POST /api/v1/instance/{name}/chat/send-presence
```

```json
{
  "chat_jid": "5511999999999@s.whatsapp.net",
  "presence": "composing"
}
```

`presence`: `composing`, `paused`, `recording`

### Block/unblock

```
POST /api/v1/instance/{name}/chat/block
```

```json
{
  "jid": "5511999999999@s.whatsapp.net",
  "action": "block"
}
```

`action`: `block`, `unblock`

### List contacts

```
GET /api/v1/instance/{name}/chat/contacts
```

### Get profile

```
GET /api/v1/instance/{name}/chat/profile/{jid}
```

---

## Group

All group endpoints are under `/api/v1/instance/{name}/group/`.

### Create group

```
POST /api/v1/instance/{name}/group/create
```

```json
{
  "name": "My Group",
  "participants": ["5511999999999"]
}
```

### List groups

```
GET /api/v1/instance/{name}/group
```

### Get group info

```
GET /api/v1/instance/{name}/group/{jid}
```

### Update group name

```
PUT /api/v1/instance/{name}/group/{jid}/name
```

```json
{ "name": "New Group Name" }
```

### Update group description

```
PUT /api/v1/instance/{name}/group/{jid}/description
```

```json
{ "description": "New description" }
```

### Update group photo

```
PUT /api/v1/instance/{name}/group/{jid}/photo
```

Multipart form with `photo` field.

### Update group settings

```
PUT /api/v1/instance/{name}/group/{jid}/settings
```

```json
{
  "locked": true,
  "announce": false
}
```

### Manage participants

```
POST /api/v1/instance/{name}/group/{jid}/participants
```

```json
{
  "action": "add",
  "participants": ["5511999999999"]
}
```

`action`: `add`, `remove`, `promote`, `demote`

### Get invite link

```
GET /api/v1/instance/{name}/group/{jid}/invite-link
```

Response `200`:

```json
{ "invite_link": "https://chat.whatsapp.com/..." }
```

### Leave group

```
DELETE /api/v1/instance/{name}/group/{jid}/leave
```

---

## Webhook

See [webhooks.md](webhooks.md) for event types and payload format.

### Set webhook

```
POST /api/v1/instance/{name}/webhook
```

```json
{
  "url": "https://my-server.com/webhook",
  "events": ["message.received", "connection.update"],
  "headers": { "X-Custom": "value" },
  "enabled": true
}
```

Empty `events` array means all events.

### Get webhook

```
GET /api/v1/instance/{name}/webhook
```

### Delete webhook

```
DELETE /api/v1/instance/{name}/webhook
```

---

## Settings

See [settings.md](settings.md) for all available settings.

### Get settings

```
GET /api/v1/instance/{name}/settings
```

### Update settings

```
PUT /api/v1/instance/{name}/settings
```

```json
{
  "always_online": true,
  "reject_call": true,
  "reject_call_message": "Can't talk right now"
}
```

---

## Error responses

All errors follow the same format:

```json
{ "error": "description of the error" }
```

Common HTTP status codes:

| Status | Meaning |
|--------|---------|
| `400` | Bad request (invalid input) |
| `401` | Unauthorized (missing or invalid token) |
| `403` | Forbidden (token doesn't have access) |
| `404` | Not found |
| `409` | Conflict (e.g. instance name already exists) |
| `503` | Instance not connected |
