CREATE TABLE instances (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name           TEXT NOT NULL UNIQUE,
    token          TEXT NOT NULL UNIQUE,
    status         TEXT NOT NULL DEFAULT 'close',
    phone          TEXT NOT NULL DEFAULT '',
    business_name  TEXT NOT NULL DEFAULT '',
    whatsmeow_jid  TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_instances_name ON instances(name);
CREATE INDEX idx_instances_token ON instances(token);

CREATE TABLE instance_settings (
    instance_id         UUID PRIMARY KEY REFERENCES instances(id) ON DELETE CASCADE,
    reject_call         BOOLEAN NOT NULL DEFAULT FALSE,
    reject_call_message TEXT NOT NULL DEFAULT '',
    groups_ignore       BOOLEAN NOT NULL DEFAULT FALSE,
    always_online       BOOLEAN NOT NULL DEFAULT FALSE,
    read_messages       BOOLEAN NOT NULL DEFAULT FALSE,
    read_receipts       BOOLEAN NOT NULL DEFAULT FALSE,
    webhook_base64      BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE webhooks (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID NOT NULL UNIQUE REFERENCES instances(id) ON DELETE CASCADE,
    url         TEXT NOT NULL,
    events      JSONB NOT NULL DEFAULT '[]',
    headers     JSONB NOT NULL DEFAULT '{}',
    enabled     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE contacts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id   UUID NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    jid           TEXT NOT NULL,
    name          TEXT NOT NULL DEFAULT '',
    push_name     TEXT NOT NULL DEFAULT '',
    business_name TEXT NOT NULL DEFAULT '',
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(instance_id, jid)
);
CREATE INDEX idx_contacts_instance ON contacts(instance_id);

CREATE TABLE messages (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id  UUID NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    message_id   TEXT NOT NULL,
    chat_jid     TEXT NOT NULL,
    sender_jid   TEXT NOT NULL,
    content_type TEXT NOT NULL DEFAULT 'text',
    content      JSONB NOT NULL DEFAULT '{}',
    timestamp    TIMESTAMPTZ NOT NULL,
    is_from_me   BOOLEAN NOT NULL DEFAULT FALSE,
    status       TEXT NOT NULL DEFAULT 'sent'
);
CREATE INDEX idx_messages_instance_chat ON messages(instance_id, chat_jid);
CREATE INDEX idx_messages_timestamp ON messages(timestamp);

CREATE TABLE chats (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id     UUID NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    jid             TEXT NOT NULL,
    name            TEXT NOT NULL DEFAULT '',
    is_group        BOOLEAN NOT NULL DEFAULT FALSE,
    unread_count    INTEGER NOT NULL DEFAULT 0,
    last_message_at TIMESTAMPTZ,
    UNIQUE(instance_id, jid)
);
CREATE INDEX idx_chats_instance ON chats(instance_id);

CREATE TABLE schema_migrations (
    version    TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
