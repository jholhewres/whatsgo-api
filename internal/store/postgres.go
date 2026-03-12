package store

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewPostgresStore(dsn string, logger *slog.Logger, maxConns, minConns int) (*PostgresStore, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parsing DSN: %w", err)
	}

	if maxConns <= 0 {
		maxConns = 50
	}
	if minConns <= 0 {
		minConns = 10
	}
	config.MaxConns = int32(maxConns)
	config.MinConns = int32(minConns)
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute
	config.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return &PostgresStore{pool: pool, logger: logger}, nil
}

func (s *PostgresStore) Pool() *pgxpool.Pool { return s.pool }

func (s *PostgresStore) Migrate(ctx context.Context) error {
	return RunPostgresMigrations(ctx, s.pool)
}

func (s *PostgresStore) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *PostgresStore) Close() error {
	s.pool.Close()
	return nil
}

// --- JSONB helpers ---

func marshalJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}

func unmarshalJSONStringSlice(data []byte) ([]string, error) {
	if data == nil {
		return nil, nil
	}
	var result []string
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func unmarshalJSONStringMap(data []byte) (map[string]string, error) {
	if data == nil {
		return nil, nil
	}
	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// --- Instances ---

const instanceColumns = `id, name, token, status, phone, business_name, whatsmeow_jid, created_at, updated_at`

func (s *PostgresStore) CreateInstance(ctx context.Context, inst *Instance) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO instances (id, name, token, status, phone, business_name, whatsmeow_jid, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		inst.ID, inst.Name, inst.Token, inst.Status, inst.Phone, inst.BusinessName,
		inst.WhatsmeowJID, inst.CreatedAt, inst.UpdatedAt,
	)
	return err
}

func (s *PostgresStore) GetInstance(ctx context.Context, id string) (*Instance, error) {
	return s.scanInstance(s.pool.QueryRow(ctx, `SELECT `+instanceColumns+` FROM instances WHERE id = $1`, id))
}

func (s *PostgresStore) GetInstanceByName(ctx context.Context, name string) (*Instance, error) {
	return s.scanInstance(s.pool.QueryRow(ctx, `SELECT `+instanceColumns+` FROM instances WHERE name = $1`, name))
}

func (s *PostgresStore) GetInstanceByToken(ctx context.Context, token string) (*Instance, error) {
	return s.scanInstance(s.pool.QueryRow(ctx, `SELECT `+instanceColumns+` FROM instances WHERE token = $1`, token))
}

func (s *PostgresStore) ListInstances(ctx context.Context) ([]*Instance, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+instanceColumns+` FROM instances ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []*Instance
	for rows.Next() {
		inst, err := s.scanInstanceRow(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}

func (s *PostgresStore) UpdateInstanceStatus(ctx context.Context, id string, status string) error {
	_, err := s.pool.Exec(ctx, `UPDATE instances SET status = $2, updated_at = NOW() WHERE id = $1`, id, status)
	return err
}

func (s *PostgresStore) UpdateInstancePhone(ctx context.Context, id string, phone, businessName, whatsmeowJID string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE instances SET phone = $2, business_name = $3, whatsmeow_jid = $4, updated_at = NOW() WHERE id = $1`,
		id, phone, businessName, whatsmeowJID,
	)
	return err
}

func (s *PostgresStore) DeleteInstance(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM instances WHERE id = $1`, id)
	return err
}

func (s *PostgresStore) scanInstance(row pgx.Row) (*Instance, error) {
	var inst Instance
	err := row.Scan(
		&inst.ID, &inst.Name, &inst.Token, &inst.Status,
		&inst.Phone, &inst.BusinessName, &inst.WhatsmeowJID,
		&inst.CreatedAt, &inst.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &inst, nil
}

func (s *PostgresStore) scanInstanceRow(rows pgx.Rows) (*Instance, error) {
	var inst Instance
	err := rows.Scan(
		&inst.ID, &inst.Name, &inst.Token, &inst.Status,
		&inst.Phone, &inst.BusinessName, &inst.WhatsmeowJID,
		&inst.CreatedAt, &inst.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &inst, nil
}

// --- Instance Settings ---

func (s *PostgresStore) GetInstanceSettings(ctx context.Context, instanceID string) (*InstanceSettings, error) {
	var settings InstanceSettings
	err := s.pool.QueryRow(ctx, `
		SELECT instance_id, reject_call, reject_call_message, groups_ignore,
		       always_online, read_messages, read_receipts, webhook_base64
		FROM instance_settings WHERE instance_id = $1`, instanceID,
	).Scan(
		&settings.InstanceID, &settings.RejectCall, &settings.RejectCallMessage,
		&settings.GroupsIgnore, &settings.AlwaysOnline, &settings.ReadMessages,
		&settings.ReadReceipts, &settings.WebhookBase64,
	)
	if err == pgx.ErrNoRows {
		return &InstanceSettings{InstanceID: instanceID}, nil
	}
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (s *PostgresStore) UpsertInstanceSettings(ctx context.Context, settings *InstanceSettings) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO instance_settings (instance_id, reject_call, reject_call_message, groups_ignore,
		                               always_online, read_messages, read_receipts, webhook_base64)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (instance_id) DO UPDATE SET
			reject_call = EXCLUDED.reject_call,
			reject_call_message = EXCLUDED.reject_call_message,
			groups_ignore = EXCLUDED.groups_ignore,
			always_online = EXCLUDED.always_online,
			read_messages = EXCLUDED.read_messages,
			read_receipts = EXCLUDED.read_receipts,
			webhook_base64 = EXCLUDED.webhook_base64`,
		settings.InstanceID, settings.RejectCall, settings.RejectCallMessage,
		settings.GroupsIgnore, settings.AlwaysOnline, settings.ReadMessages,
		settings.ReadReceipts, settings.WebhookBase64,
	)
	return err
}

// --- Webhooks ---

func (s *PostgresStore) CreateWebhook(ctx context.Context, wh *Webhook) error {
	eventsJSON, err := marshalJSON(wh.Events)
	if err != nil {
		return fmt.Errorf("marshaling events: %w", err)
	}
	headersJSON, err := marshalJSON(wh.Headers)
	if err != nil {
		return fmt.Errorf("marshaling headers: %w", err)
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO webhooks (id, instance_id, url, events, headers, enabled, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		wh.ID, wh.InstanceID, wh.URL, eventsJSON, headersJSON, wh.Enabled, wh.CreatedAt,
	)
	return err
}

func (s *PostgresStore) GetWebhookByInstance(ctx context.Context, instanceID string) (*Webhook, error) {
	var wh Webhook
	var eventsJSON, headersJSON []byte
	err := s.pool.QueryRow(ctx, `
		SELECT id, instance_id, url, events, headers, enabled, created_at
		FROM webhooks WHERE instance_id = $1`, instanceID,
	).Scan(&wh.ID, &wh.InstanceID, &wh.URL, &eventsJSON, &headersJSON, &wh.Enabled, &wh.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	wh.Events, _ = unmarshalJSONStringSlice(eventsJSON)
	wh.Headers, _ = unmarshalJSONStringMap(headersJSON)
	return &wh, nil
}

func (s *PostgresStore) UpdateWebhook(ctx context.Context, wh *Webhook) error {
	eventsJSON, err := marshalJSON(wh.Events)
	if err != nil {
		return fmt.Errorf("marshaling events: %w", err)
	}
	headersJSON, err := marshalJSON(wh.Headers)
	if err != nil {
		return fmt.Errorf("marshaling headers: %w", err)
	}
	_, err = s.pool.Exec(ctx, `
		UPDATE webhooks SET url = $2, events = $3, headers = $4, enabled = $5 WHERE instance_id = $1`,
		wh.InstanceID, wh.URL, eventsJSON, headersJSON, wh.Enabled,
	)
	return err
}

func (s *PostgresStore) DeleteWebhook(ctx context.Context, instanceID string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM webhooks WHERE instance_id = $1`, instanceID)
	return err
}

// --- Contacts ---

func (s *PostgresStore) UpsertContact(ctx context.Context, contact *Contact) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO contacts (id, instance_id, jid, name, push_name, business_name, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (instance_id, jid) DO UPDATE SET
			name = EXCLUDED.name,
			push_name = EXCLUDED.push_name,
			business_name = EXCLUDED.business_name,
			updated_at = EXCLUDED.updated_at`,
		contact.ID, contact.InstanceID, contact.JID, contact.Name,
		contact.PushName, contact.BusinessName, contact.UpdatedAt,
	)
	return err
}

func (s *PostgresStore) ListContacts(ctx context.Context, instanceID string) ([]*Contact, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, instance_id, jid, name, push_name, business_name, updated_at
		FROM contacts WHERE instance_id = $1 ORDER BY push_name`, instanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []*Contact
	for rows.Next() {
		var c Contact
		if err := rows.Scan(&c.ID, &c.InstanceID, &c.JID, &c.Name, &c.PushName, &c.BusinessName, &c.UpdatedAt); err != nil {
			return nil, err
		}
		contacts = append(contacts, &c)
	}
	return contacts, rows.Err()
}

func (s *PostgresStore) GetContact(ctx context.Context, instanceID, jid string) (*Contact, error) {
	var c Contact
	err := s.pool.QueryRow(ctx, `
		SELECT id, instance_id, jid, name, push_name, business_name, updated_at
		FROM contacts WHERE instance_id = $1 AND jid = $2`, instanceID, jid,
	).Scan(&c.ID, &c.InstanceID, &c.JID, &c.Name, &c.PushName, &c.BusinessName, &c.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// --- Messages ---

func (s *PostgresStore) SaveMessage(ctx context.Context, msg *Message) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO messages (id, instance_id, message_id, chat_jid, sender_jid, content_type, content, timestamp, is_from_me, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		msg.ID, msg.InstanceID, msg.MessageID, msg.ChatJID, msg.SenderJID,
		msg.ContentType, msg.Content, msg.Timestamp, msg.IsFromMe, msg.Status,
	)
	return err
}

func (s *PostgresStore) GetMessages(ctx context.Context, instanceID, chatJID string, limit int) ([]*Message, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, instance_id, message_id, chat_jid, sender_jid, content_type, content, timestamp, is_from_me, status
		FROM messages WHERE instance_id = $1 AND chat_jid = $2
		ORDER BY timestamp DESC LIMIT $3`, instanceID, chatJID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.InstanceID, &m.MessageID, &m.ChatJID, &m.SenderJID,
			&m.ContentType, &m.Content, &m.Timestamp, &m.IsFromMe, &m.Status); err != nil {
			return nil, err
		}
		messages = append(messages, &m)
	}
	return messages, rows.Err()
}

// --- Chats ---

func (s *PostgresStore) UpsertChat(ctx context.Context, chat *Chat) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO chats (id, instance_id, jid, name, is_group, unread_count, last_message_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (instance_id, jid) DO UPDATE SET
			name = EXCLUDED.name,
			is_group = EXCLUDED.is_group,
			unread_count = EXCLUDED.unread_count,
			last_message_at = EXCLUDED.last_message_at`,
		chat.ID, chat.InstanceID, chat.JID, chat.Name, chat.IsGroup, chat.UnreadCount, chat.LastMessageAt,
	)
	return err
}

func (s *PostgresStore) ListChats(ctx context.Context, instanceID string) ([]*Chat, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, instance_id, jid, name, is_group, unread_count, last_message_at
		FROM chats WHERE instance_id = $1 ORDER BY last_message_at DESC NULLS LAST`, instanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*Chat
	for rows.Next() {
		var c Chat
		if err := rows.Scan(&c.ID, &c.InstanceID, &c.JID, &c.Name, &c.IsGroup, &c.UnreadCount, &c.LastMessageAt); err != nil {
			return nil, err
		}
		chats = append(chats, &c)
	}
	return chats, rows.Err()
}
