CREATE TABLE instances (
    id             TEXT PRIMARY KEY,
    name           TEXT NOT NULL UNIQUE,
    token          TEXT NOT NULL UNIQUE,
    status         TEXT NOT NULL DEFAULT 'close',
    phone          TEXT NOT NULL DEFAULT '',
    business_name  TEXT NOT NULL DEFAULT '',
    whatsmeow_jid  TEXT NOT NULL DEFAULT '',
    created_at     TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at     TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_instances_name ON instances(name);
CREATE INDEX idx_instances_token ON instances(token);

CREATE TABLE instance_settings (
    instance_id         TEXT PRIMARY KEY REFERENCES instances(id) ON DELETE CASCADE,
    reject_call         INTEGER NOT NULL DEFAULT 0,
    reject_call_message TEXT NOT NULL DEFAULT '',
    groups_ignore       INTEGER NOT NULL DEFAULT 0,
    always_online       INTEGER NOT NULL DEFAULT 0,
    read_messages       INTEGER NOT NULL DEFAULT 0,
    read_receipts       INTEGER NOT NULL DEFAULT 0,
    webhook_base64      INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE webhooks (
    id          TEXT PRIMARY KEY,
    instance_id TEXT NOT NULL UNIQUE REFERENCES instances(id) ON DELETE CASCADE,
    url         TEXT NOT NULL,
    events      TEXT NOT NULL DEFAULT '[]',
    headers     TEXT NOT NULL DEFAULT '{}',
    enabled     INTEGER NOT NULL DEFAULT 1,
    created_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE contacts (
    id            TEXT PRIMARY KEY,
    instance_id   TEXT NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    jid           TEXT NOT NULL,
    name          TEXT NOT NULL DEFAULT '',
    push_name     TEXT NOT NULL DEFAULT '',
    business_name TEXT NOT NULL DEFAULT '',
    updated_at    TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(instance_id, jid)
);
CREATE INDEX idx_contacts_instance ON contacts(instance_id);

CREATE TABLE messages (
    id           TEXT PRIMARY KEY,
    instance_id  TEXT NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    message_id   TEXT NOT NULL,
    chat_jid     TEXT NOT NULL,
    sender_jid   TEXT NOT NULL,
    content_type TEXT NOT NULL DEFAULT 'text',
    content      TEXT NOT NULL DEFAULT '{}',
    timestamp    TEXT NOT NULL,
    is_from_me   INTEGER NOT NULL DEFAULT 0,
    status       TEXT NOT NULL DEFAULT 'sent'
);
CREATE INDEX idx_messages_instance_chat ON messages(instance_id, chat_jid);
CREATE INDEX idx_messages_timestamp ON messages(timestamp);

CREATE TABLE chats (
    id              TEXT PRIMARY KEY,
    instance_id     TEXT NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    jid             TEXT NOT NULL,
    name            TEXT NOT NULL DEFAULT '',
    is_group        INTEGER NOT NULL DEFAULT 0,
    unread_count    INTEGER NOT NULL DEFAULT 0,
    last_message_at TEXT,
    UNIQUE(instance_id, jid)
);
CREATE INDEX idx_chats_instance ON chats(instance_id);

CREATE TABLE schema_migrations (
    version    TEXT PRIMARY KEY,
    applied_at TEXT NOT NULL DEFAULT (datetime('now'))
);
