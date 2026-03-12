package store

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jholhewres/whatsgo-api/internal/config"
)

type Store interface {
	// Instances
	CreateInstance(ctx context.Context, inst *Instance) error
	GetInstance(ctx context.Context, id string) (*Instance, error)
	GetInstanceByName(ctx context.Context, name string) (*Instance, error)
	GetInstanceByToken(ctx context.Context, token string) (*Instance, error)
	ListInstances(ctx context.Context) ([]*Instance, error)
	UpdateInstanceStatus(ctx context.Context, id string, status string) error
	UpdateInstancePhone(ctx context.Context, id string, phone, businessName, whatsmeowJID string) error
	DeleteInstance(ctx context.Context, id string) error

	// Instance Settings
	GetInstanceSettings(ctx context.Context, instanceID string) (*InstanceSettings, error)
	UpsertInstanceSettings(ctx context.Context, settings *InstanceSettings) error

	// Webhooks
	CreateWebhook(ctx context.Context, wh *Webhook) error
	GetWebhookByInstance(ctx context.Context, instanceID string) (*Webhook, error)
	UpdateWebhook(ctx context.Context, wh *Webhook) error
	DeleteWebhook(ctx context.Context, instanceID string) error

	// Contacts
	UpsertContact(ctx context.Context, contact *Contact) error
	ListContacts(ctx context.Context, instanceID string) ([]*Contact, error)
	GetContact(ctx context.Context, instanceID, jid string) (*Contact, error)

	// Messages
	SaveMessage(ctx context.Context, msg *Message) error
	GetMessages(ctx context.Context, instanceID, chatJID string, limit int) ([]*Message, error)

	// Chats
	UpsertChat(ctx context.Context, chat *Chat) error
	ListChats(ctx context.Context, instanceID string) ([]*Chat, error)

	// Lifecycle
	Migrate(ctx context.Context) error
	Ping(ctx context.Context) error
	Close() error
}

// Open creates a new Store based on the database backend config.
func Open(cfg config.DatabaseConfig, logger *slog.Logger) (Store, error) {
	switch cfg.Backend {
	case "postgresql", "postgres", "":
		return NewPostgresStore(cfg.PostgreSQL.DSN(), logger, cfg.PostgreSQL.MaxConns, cfg.PostgreSQL.MinConns)
	case "sqlite":
		return NewSQLiteStore(cfg.SQLite.Path, logger)
	default:
		return nil, fmt.Errorf("unsupported database backend: %s", cfg.Backend)
	}
}

// Domain types

type Instance struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Token         string    `json:"token,omitempty"`
	Status        string    `json:"status"`
	Phone         string    `json:"phone,omitempty"`
	BusinessName  string    `json:"business_name,omitempty"`
	WhatsmeowJID  string    `json:"whatsmeow_jid,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type InstanceSettings struct {
	InstanceID        string `json:"instance_id"`
	RejectCall        bool   `json:"reject_call"`
	RejectCallMessage string `json:"reject_call_message"`
	GroupsIgnore      bool   `json:"groups_ignore"`
	AlwaysOnline      bool   `json:"always_online"`
	ReadMessages      bool   `json:"read_messages"`
	ReadReceipts      bool   `json:"read_receipts"`
	WebhookBase64     bool   `json:"webhook_base64"`
}

type Webhook struct {
	ID         string            `json:"id"`
	InstanceID string            `json:"instance_id"`
	URL        string            `json:"url"`
	Events     []string          `json:"events"`
	Headers    map[string]string `json:"headers"`
	Enabled    bool              `json:"enabled"`
	CreatedAt  time.Time         `json:"created_at"`
}

type Contact struct {
	ID           string    `json:"id"`
	InstanceID   string    `json:"instance_id"`
	JID          string    `json:"jid"`
	Name         string    `json:"name"`
	PushName     string    `json:"push_name"`
	BusinessName string    `json:"business_name"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Message struct {
	ID          string          `json:"id"`
	InstanceID  string          `json:"instance_id"`
	MessageID   string          `json:"message_id"`
	ChatJID     string          `json:"chat_jid"`
	SenderJID   string          `json:"sender_jid"`
	ContentType string          `json:"content_type"`
	Content     json.RawMessage `json:"content"`
	Timestamp   time.Time       `json:"timestamp"`
	IsFromMe    bool            `json:"is_from_me"`
	Status      string          `json:"status"`
}

type Chat struct {
	ID            string     `json:"id"`
	InstanceID    string     `json:"instance_id"`
	JID           string     `json:"jid"`
	Name          string     `json:"name"`
	IsGroup       bool       `json:"is_group"`
	UnreadCount   int        `json:"unread_count"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
}
