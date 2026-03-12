package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// parseSQLiteTime parses time from SQLite which may be RFC3339 or "YYYY-MM-DD HH:MM:SS" format.
func parseSQLiteTime(s string) time.Time {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return t
	}
	return time.Time{}
}

type SQLiteStore struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewSQLiteStore(path string, logger *slog.Logger) (*SQLiteStore, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("creating data directory: %w", err)
	}

	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("opening sqlite: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("pinging sqlite: %w", err)
	}

	db.SetMaxOpenConns(1)

	return &SQLiteStore{db: db, logger: logger}, nil
}

func (s *SQLiteStore) DB() *sql.DB { return s.db }

func (s *SQLiteStore) Migrate(ctx context.Context) error {
	return RunSQLiteMigrations(ctx, s.db)
}

func (s *SQLiteStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// --- Instances ---

func (s *SQLiteStore) CreateInstance(ctx context.Context, inst *Instance) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO instances (id, name, token, status, phone, business_name, whatsmeow_jid, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		inst.ID, inst.Name, inst.Token, inst.Status, inst.Phone, inst.BusinessName,
		inst.WhatsmeowJID, inst.CreatedAt.Format(time.RFC3339), inst.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

func (s *SQLiteStore) GetInstance(ctx context.Context, id string) (*Instance, error) {
	return s.scanInstance(s.db.QueryRowContext(ctx,
		`SELECT id, name, token, status, phone, business_name, whatsmeow_jid, created_at, updated_at FROM instances WHERE id = ?`, id))
}

func (s *SQLiteStore) GetInstanceByName(ctx context.Context, name string) (*Instance, error) {
	return s.scanInstance(s.db.QueryRowContext(ctx,
		`SELECT id, name, token, status, phone, business_name, whatsmeow_jid, created_at, updated_at FROM instances WHERE name = ?`, name))
}

func (s *SQLiteStore) GetInstanceByToken(ctx context.Context, token string) (*Instance, error) {
	return s.scanInstance(s.db.QueryRowContext(ctx,
		`SELECT id, name, token, status, phone, business_name, whatsmeow_jid, created_at, updated_at FROM instances WHERE token = ?`, token))
}

func (s *SQLiteStore) ListInstances(ctx context.Context) ([]*Instance, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, token, status, phone, business_name, whatsmeow_jid, created_at, updated_at FROM instances ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []*Instance
	for rows.Next() {
		inst, err := s.scanInstanceRows(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}

func (s *SQLiteStore) UpdateInstanceStatus(ctx context.Context, id string, status string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE instances SET status = ?, updated_at = datetime('now') WHERE id = ?`, status, id)
	return err
}

func (s *SQLiteStore) UpdateInstancePhone(ctx context.Context, id string, phone, businessName, whatsmeowJID string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE instances SET phone = ?, business_name = ?, whatsmeow_jid = ?, updated_at = datetime('now') WHERE id = ?`,
		phone, businessName, whatsmeowJID, id,
	)
	return err
}

func (s *SQLiteStore) DeleteInstance(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM instances WHERE id = ?`, id)
	return err
}

func (s *SQLiteStore) scanInstance(row *sql.Row) (*Instance, error) {
	var inst Instance
	var createdAt, updatedAt string
	err := row.Scan(
		&inst.ID, &inst.Name, &inst.Token, &inst.Status,
		&inst.Phone, &inst.BusinessName, &inst.WhatsmeowJID,
		&createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	inst.CreatedAt = parseSQLiteTime(createdAt)
	inst.UpdatedAt = parseSQLiteTime(updatedAt)
	return &inst, nil
}

func (s *SQLiteStore) scanInstanceRows(rows *sql.Rows) (*Instance, error) {
	var inst Instance
	var createdAt, updatedAt string
	err := rows.Scan(
		&inst.ID, &inst.Name, &inst.Token, &inst.Status,
		&inst.Phone, &inst.BusinessName, &inst.WhatsmeowJID,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}
	inst.CreatedAt = parseSQLiteTime(createdAt)
	inst.UpdatedAt = parseSQLiteTime(updatedAt)
	return &inst, nil
}

// --- Instance Settings ---

func (s *SQLiteStore) GetInstanceSettings(ctx context.Context, instanceID string) (*InstanceSettings, error) {
	var settings InstanceSettings
	err := s.db.QueryRowContext(ctx, `
		SELECT instance_id, reject_call, reject_call_message, groups_ignore,
		       always_online, read_messages, read_receipts, webhook_base64
		FROM instance_settings WHERE instance_id = ?`, instanceID,
	).Scan(
		&settings.InstanceID, &settings.RejectCall, &settings.RejectCallMessage,
		&settings.GroupsIgnore, &settings.AlwaysOnline, &settings.ReadMessages,
		&settings.ReadReceipts, &settings.WebhookBase64,
	)
	if err == sql.ErrNoRows {
		return &InstanceSettings{InstanceID: instanceID}, nil
	}
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (s *SQLiteStore) UpsertInstanceSettings(ctx context.Context, settings *InstanceSettings) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO instance_settings (instance_id, reject_call, reject_call_message, groups_ignore,
		                               always_online, read_messages, read_receipts, webhook_base64)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (instance_id) DO UPDATE SET
			reject_call = excluded.reject_call,
			reject_call_message = excluded.reject_call_message,
			groups_ignore = excluded.groups_ignore,
			always_online = excluded.always_online,
			read_messages = excluded.read_messages,
			read_receipts = excluded.read_receipts,
			webhook_base64 = excluded.webhook_base64`,
		settings.InstanceID, settings.RejectCall, settings.RejectCallMessage,
		settings.GroupsIgnore, settings.AlwaysOnline, settings.ReadMessages,
		settings.ReadReceipts, settings.WebhookBase64,
	)
	return err
}

// --- Webhooks ---

func (s *SQLiteStore) CreateWebhook(ctx context.Context, wh *Webhook) error {
	eventsJSON, _ := json.Marshal(wh.Events)
	headersJSON, _ := json.Marshal(wh.Headers)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO webhooks (id, instance_id, url, events, headers, enabled, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		wh.ID, wh.InstanceID, wh.URL, string(eventsJSON), string(headersJSON),
		wh.Enabled, wh.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (s *SQLiteStore) GetWebhookByInstance(ctx context.Context, instanceID string) (*Webhook, error) {
	var wh Webhook
	var eventsStr, headersStr, createdAt string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, instance_id, url, events, headers, enabled, created_at
		FROM webhooks WHERE instance_id = ?`, instanceID,
	).Scan(&wh.ID, &wh.InstanceID, &wh.URL, &eventsStr, &headersStr, &wh.Enabled, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(eventsStr), &wh.Events)
	json.Unmarshal([]byte(headersStr), &wh.Headers)
	wh.CreatedAt = parseSQLiteTime(createdAt)
	return &wh, nil
}

func (s *SQLiteStore) UpdateWebhook(ctx context.Context, wh *Webhook) error {
	eventsJSON, _ := json.Marshal(wh.Events)
	headersJSON, _ := json.Marshal(wh.Headers)
	_, err := s.db.ExecContext(ctx, `
		UPDATE webhooks SET url = ?, events = ?, headers = ?, enabled = ? WHERE instance_id = ?`,
		wh.URL, string(eventsJSON), string(headersJSON), wh.Enabled, wh.InstanceID,
	)
	return err
}

func (s *SQLiteStore) DeleteWebhook(ctx context.Context, instanceID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM webhooks WHERE instance_id = ?`, instanceID)
	return err
}

// --- Contacts ---

func (s *SQLiteStore) UpsertContact(ctx context.Context, contact *Contact) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO contacts (id, instance_id, jid, name, push_name, business_name, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (instance_id, jid) DO UPDATE SET
			name = excluded.name,
			push_name = excluded.push_name,
			business_name = excluded.business_name,
			updated_at = excluded.updated_at`,
		contact.ID, contact.InstanceID, contact.JID, contact.Name,
		contact.PushName, contact.BusinessName, contact.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

func (s *SQLiteStore) ListContacts(ctx context.Context, instanceID string) ([]*Contact, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, instance_id, jid, name, push_name, business_name, updated_at
		FROM contacts WHERE instance_id = ? ORDER BY push_name`, instanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []*Contact
	for rows.Next() {
		var c Contact
		var updatedAt string
		if err := rows.Scan(&c.ID, &c.InstanceID, &c.JID, &c.Name, &c.PushName, &c.BusinessName, &updatedAt); err != nil {
			return nil, err
		}
		c.UpdatedAt = parseSQLiteTime(updatedAt)
		contacts = append(contacts, &c)
	}
	return contacts, rows.Err()
}

func (s *SQLiteStore) GetContact(ctx context.Context, instanceID, jid string) (*Contact, error) {
	var c Contact
	var updatedAt string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, instance_id, jid, name, push_name, business_name, updated_at
		FROM contacts WHERE instance_id = ? AND jid = ?`, instanceID, jid,
	).Scan(&c.ID, &c.InstanceID, &c.JID, &c.Name, &c.PushName, &c.BusinessName, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &c, nil
}

// --- Messages ---

func (s *SQLiteStore) SaveMessage(ctx context.Context, msg *Message) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO messages (id, instance_id, message_id, chat_jid, sender_jid, content_type, content, timestamp, is_from_me, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		msg.ID, msg.InstanceID, msg.MessageID, msg.ChatJID, msg.SenderJID,
		msg.ContentType, string(msg.Content), msg.Timestamp.Format(time.RFC3339), msg.IsFromMe, msg.Status,
	)
	return err
}

func (s *SQLiteStore) GetMessages(ctx context.Context, instanceID, chatJID string, limit int) ([]*Message, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, instance_id, message_id, chat_jid, sender_jid, content_type, content, timestamp, is_from_me, status
		FROM messages WHERE instance_id = ? AND chat_jid = ?
		ORDER BY timestamp DESC LIMIT ?`, instanceID, chatJID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var m Message
		var contentStr, ts string
		if err := rows.Scan(&m.ID, &m.InstanceID, &m.MessageID, &m.ChatJID, &m.SenderJID,
			&m.ContentType, &contentStr, &ts, &m.IsFromMe, &m.Status); err != nil {
			return nil, err
		}
		m.Content = json.RawMessage(contentStr)
		m.Timestamp = parseSQLiteTime(ts)
		messages = append(messages, &m)
	}
	return messages, rows.Err()
}

// --- Chats ---

func (s *SQLiteStore) UpsertChat(ctx context.Context, chat *Chat) error {
	var lastMsg *string
	if chat.LastMessageAt != nil {
		t := chat.LastMessageAt.Format(time.RFC3339)
		lastMsg = &t
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO chats (id, instance_id, jid, name, is_group, unread_count, last_message_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (instance_id, jid) DO UPDATE SET
			name = excluded.name,
			is_group = excluded.is_group,
			unread_count = excluded.unread_count,
			last_message_at = excluded.last_message_at`,
		chat.ID, chat.InstanceID, chat.JID, chat.Name, chat.IsGroup, chat.UnreadCount, lastMsg,
	)
	return err
}

func (s *SQLiteStore) ListChats(ctx context.Context, instanceID string) ([]*Chat, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, instance_id, jid, name, is_group, unread_count, last_message_at
		FROM chats WHERE instance_id = ? ORDER BY last_message_at DESC`, instanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*Chat
	for rows.Next() {
		var c Chat
		var lastMsg *string
		if err := rows.Scan(&c.ID, &c.InstanceID, &c.JID, &c.Name, &c.IsGroup, &c.UnreadCount, &lastMsg); err != nil {
			return nil, err
		}
		if lastMsg != nil {
			t := parseSQLiteTime(*lastMsg)
			c.LastMessageAt = &t
		}
		chats = append(chats, &c)
	}
	return chats, rows.Err()
}
