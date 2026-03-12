# WhatsGo API

API REST para WhatsApp multi-instancia, construida em Go com [whatsmeow](https://github.com/tulir/whatsmeow). Simples, direta, e pronta para producao.

## Funcionalidades

- **Multi-instancia** - Gerencie multiplas conexoes WhatsApp simultaneamente
- **Autenticacao dupla** - API key global + token por instancia
- **QR Code e Pairing Code** - Duas formas de conectar
- **Mensagens** - Texto, midia, localizacao, contato, sticker, reacao
- **Grupos** - CRUD completo, participantes, convites
- **Webhooks** - Por instancia + global, com retry exponencial
- **Settings por instancia** - Rejeitar chamadas, always online, auto-read, etc.
- **Frontend embutido** - Dashboard React para gerenciamento
- **PostgreSQL + SQLite** - Postgres para producao, SQLite para dev local

## Quick Start

### Docker (recomendado)

```bash
git clone https://github.com/jholhewres/whatsgo-api.git
cd whatsgo-api
docker compose up -d
```

A API estara disponivel em `http://localhost:8550`.

### Build manual

**Requisitos:** Go 1.24+, Node.js 22+, GCC (para SQLite)

```bash
# Clonar
git clone https://github.com/jholhewres/whatsgo-api.git
cd whatsgo-api

# Build (frontend + backend)
make build

# Executar
./whatsgo
```

### Desenvolvimento

```bash
# Backend
make dev

# Frontend (em outro terminal)
make dev-frontend
```

## Configuracao

Crie um arquivo `whatsgo.yaml` (ou defina `WHATSGO_CONFIG` para outro caminho):

```yaml
server:
  host: "0.0.0.0"
  port: 8550
  base_url: "http://localhost:8550"

database:
  backend: "postgresql"   # ou "sqlite"
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
  format: "text"   # ou "json"
```

Variaveis de ambiente sao suportadas no formato `${VAR:-default}`.

## Autenticacao

Toda requisicao precisa de um token. Dois tipos:

| Tipo | Header | Escopo |
|------|--------|--------|
| API Key Global | `X-API-Key: <key>` | Acesso total a todas instancias |
| Token de Instancia | `Authorization: Bearer <token>` | Acesso somente a instancia do token |

O token de instancia e retornado ao criar uma instancia.

## Endpoints

### Instance

| Metodo | Path | Descricao |
|--------|------|-----------|
| `POST` | `/api/v1/instance/create` | Criar instancia |
| `POST` | `/api/v1/instance/{name}/connect` | Conectar (QR ou pairing code) |
| `POST` | `/api/v1/instance/{name}/restart` | Reiniciar conexao |
| `GET` | `/api/v1/instance/{name}/status` | Status da conexao |
| `GET` | `/api/v1/instance` | Listar instancias |
| `DELETE` | `/api/v1/instance/{name}/logout` | Logout do WhatsApp |
| `DELETE` | `/api/v1/instance/{name}` | Deletar instancia |

### Messages

| Metodo | Path | Descricao |
|--------|------|-----------|
| `POST` | `/api/v1/instance/{name}/message/send-text` | Enviar texto |
| `POST` | `/api/v1/instance/{name}/message/send-media` | Enviar midia |
| `POST` | `/api/v1/instance/{name}/message/send-location` | Enviar localizacao |
| `POST` | `/api/v1/instance/{name}/message/send-contact` | Enviar contato |
| `POST` | `/api/v1/instance/{name}/message/send-reaction` | Enviar reacao |
| `POST` | `/api/v1/instance/{name}/message/send-sticker` | Enviar sticker |

### Chat

| Metodo | Path | Descricao |
|--------|------|-----------|
| `POST` | `/api/v1/instance/{name}/chat/check-number` | Verificar numeros no WhatsApp |
| `POST` | `/api/v1/instance/{name}/chat/mark-read` | Marcar como lido |
| `POST` | `/api/v1/instance/{name}/chat/delete-message` | Deletar mensagem |
| `POST` | `/api/v1/instance/{name}/chat/edit-message` | Editar mensagem |
| `POST` | `/api/v1/instance/{name}/chat/send-presence` | Digitando/gravando |
| `POST` | `/api/v1/instance/{name}/chat/block` | Bloquear/desbloquear |
| `GET` | `/api/v1/instance/{name}/chat/contacts` | Listar contatos |
| `GET` | `/api/v1/instance/{name}/chat/profile/{jid}` | Perfil do usuario |

### Group

| Metodo | Path | Descricao |
|--------|------|-----------|
| `POST` | `/api/v1/instance/{name}/group/create` | Criar grupo |
| `GET` | `/api/v1/instance/{name}/group` | Listar grupos |
| `GET` | `/api/v1/instance/{name}/group/{jid}` | Info do grupo |
| `PUT` | `/api/v1/instance/{name}/group/{jid}/name` | Atualizar nome |
| `PUT` | `/api/v1/instance/{name}/group/{jid}/description` | Atualizar descricao |
| `PUT` | `/api/v1/instance/{name}/group/{jid}/photo` | Atualizar foto |
| `PUT` | `/api/v1/instance/{name}/group/{jid}/settings` | Atualizar config |
| `POST` | `/api/v1/instance/{name}/group/{jid}/participants` | Gerenciar participantes |
| `GET` | `/api/v1/instance/{name}/group/{jid}/invite-link` | Link de convite |
| `DELETE` | `/api/v1/instance/{name}/group/{jid}/leave` | Sair do grupo |

### Webhook

| Metodo | Path | Descricao |
|--------|------|-----------|
| `POST` | `/api/v1/instance/{name}/webhook` | Configurar webhook |
| `GET` | `/api/v1/instance/{name}/webhook` | Obter config |
| `DELETE` | `/api/v1/instance/{name}/webhook` | Remover webhook |

### Settings

| Metodo | Path | Descricao |
|--------|------|-----------|
| `GET` | `/api/v1/instance/{name}/settings` | Obter settings |
| `PUT` | `/api/v1/instance/{name}/settings` | Atualizar settings |

## Exemplos

### Criar instancia

```bash
curl -X POST http://localhost:8550/api/v1/instance/create \
  -H "X-API-Key: sua-chave" \
  -H "Content-Type: application/json" \
  -d '{"name": "minha-instancia"}'
```

### Conectar via QR Code

```bash
curl -X POST http://localhost:8550/api/v1/instance/minha-instancia/connect \
  -H "Authorization: Bearer <token-da-instancia>"
```

### Conectar via Pairing Code

```bash
curl -X POST http://localhost:8550/api/v1/instance/minha-instancia/connect \
  -H "Authorization: Bearer <token-da-instancia>" \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+5511999999999"}'
```

### Enviar mensagem

```bash
curl -X POST http://localhost:8550/api/v1/instance/minha-instancia/message/send-text \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"number": "5511999999999", "text": "Ola!"}'
```

### Configurar webhook

```bash
curl -X POST http://localhost:8550/api/v1/instance/minha-instancia/webhook \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://meu-servidor.com/webhook",
    "events": ["message.received", "connection.update"],
    "enabled": true
  }'
```

## Eventos de Webhook

| Evento | Descricao |
|--------|-----------|
| `connection.update` | Mudanca de estado (open/close/connecting) |
| `message.received` | Mensagem recebida |
| `message.sent` | Mensagem enviada por este dispositivo |
| `message.updated` | Status da mensagem (delivered/read) |
| `message.deleted` | Mensagem deletada |
| `message.reaction` | Reacao recebida |
| `presence.update` | Online/offline/digitando |
| `group.update` | Metadados do grupo alterados |
| `group.participants` | Membro entrou/saiu/promovido/rebaixado |
| `call.received` | Chamada recebida |
| `qrcode.updated` | Novo QR code gerado |

Payload:

```json
{
  "event": "message.received",
  "instance": "minha-instancia",
  "data": { ... },
  "timestamp": "2026-03-12T15:04:05Z",
  "server_url": "http://localhost:8550"
}
```

## Settings por Instancia

| Setting | Tipo | Descricao |
|---------|------|-----------|
| `reject_call` | bool | Auto-rejeitar chamadas |
| `reject_call_message` | string | Mensagem ao rejeitar |
| `groups_ignore` | bool | Ignorar eventos de grupo |
| `always_online` | bool | Manter presenca online |
| `read_messages` | bool | Auto-marcar como lido |
| `read_receipts` | bool | Enviar recibos de leitura |
| `webhook_base64` | bool | Midia como base64 nos webhooks |

## Estrutura do Projeto

```
cmd/whatsgo/          # Entry point
internal/
  api/                # Handlers HTTP e tipos
  auth/               # Middleware de autenticacao
  cache/              # Cache generico com TTL
  config/             # Sistema de configuracao YAML
  server/             # Servidor HTTP e middlewares
  store/              # Interface de persistencia (PostgreSQL + SQLite)
  webhook/            # Dispatcher de webhooks com retry
  whatsapp/           # Manager de instancias WhatsApp (whatsmeow)
  webui/              # Frontend React embutido
web/                  # Codigo fonte do frontend React
```

## Contribuindo

Contribuicoes sao bem-vindas! Abra uma issue ou pull request.

## Licenca

Este projeto esta licenciado sob a [MIT License](LICENSE).
