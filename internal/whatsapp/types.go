package whatsapp

import "time"

// ConnectionState represents the state of a WhatsApp connection.
type ConnectionState string

const (
	StateOpen       ConnectionState = "open"
	StateClose      ConnectionState = "close"
	StateConnecting ConnectionState = "connecting"
)

// EventType represents the type of event dispatched via webhooks.
type EventType string

const (
	EventConnectionUpdate  EventType = "connection.update"
	EventMessageReceived   EventType = "message.received"
	EventMessageSent       EventType = "message.sent"
	EventMessageUpdated    EventType = "message.updated"
	EventMessageDeleted    EventType = "message.deleted"
	EventMessageReaction   EventType = "message.reaction"
	EventPresenceUpdate    EventType = "presence.update"
	EventGroupUpdate       EventType = "group.update"
	EventGroupParticipants EventType = "group.participants"
	EventCallReceived      EventType = "call.received"
	EventQRCodeUpdated     EventType = "qrcode.updated"
)

// Event is the unified event payload dispatched to webhooks.
type Event struct {
	Event     EventType `json:"event"`
	Instance  string    `json:"instance"`
	Data      any       `json:"data"`
	Timestamp time.Time `json:"timestamp"`
	ServerURL string    `json:"server_url"`
}

// QREvent represents a QR code event from the login flow.
type QREvent struct {
	Type string // "code", "success", "timeout", "error"
	Code string // QR code string (when Type == "code")
}

// MessageSendResponse is returned after sending a message.
type MessageSendResponse struct {
	MessageID string
	Timestamp time.Time
}

// ConnectionUpdate is the data payload for connection.update events.
type ConnectionUpdate struct {
	State  string `json:"state"`
	Reason string `json:"reason,omitempty"`
}

// MessagePayload is the data payload for message events.
type MessagePayload struct {
	MessageID   string    `json:"message_id"`
	ChatJID     string    `json:"chat_jid"`
	SenderJID   string    `json:"sender_jid"`
	SenderName  string    `json:"sender_name,omitempty"`
	IsGroup     bool      `json:"is_group"`
	IsFromMe    bool      `json:"is_from_me"`
	ContentType string    `json:"content_type"`
	Content     any       `json:"content"`
	Timestamp   time.Time `json:"timestamp"`
}

// ReceiptPayload is the data payload for message.updated events.
type ReceiptPayload struct {
	MessageIDs []string  `json:"message_ids"`
	ChatJID    string    `json:"chat_jid"`
	SenderJID  string    `json:"sender_jid"`
	Type       string    `json:"type"` // delivered, read, played
	Timestamp  time.Time `json:"timestamp"`
}

// PresencePayload is the data payload for presence.update events.
type PresencePayload struct {
	JID         string     `json:"jid"`
	Available   bool       `json:"available"`
	LastSeen    *time.Time `json:"last_seen,omitempty"`
	ChatPresence string   `json:"chat_presence,omitempty"` // composing, paused
}

// GroupUpdatePayload is the data payload for group.update events.
type GroupUpdatePayload struct {
	JID       string `json:"jid"`
	Name      string `json:"name,omitempty"`
	Topic     string `json:"topic,omitempty"`
	SenderJID string `json:"sender_jid"`
}

// GroupParticipantsPayload is the data payload for group.participants events.
type GroupParticipantsPayload struct {
	JID          string   `json:"jid"`
	Action       string   `json:"action"` // join, leave, promote, demote
	Participants []string `json:"participants"`
	SenderJID    string   `json:"sender_jid"`
}

// CallPayload is the data payload for call.received events.
type CallPayload struct {
	CallerJID string `json:"caller_jid"`
	CallID    string `json:"call_id"`
	Type      string `json:"type"` // audio, video
	IsGroup   bool   `json:"is_group"`
}

// ProfileInfo is returned by GetProfile.
type ProfileInfo struct {
	JID          string `json:"jid"`
	Name         string `json:"name,omitempty"`
	Status       string `json:"status,omitempty"`
	PictureURL   string `json:"picture_url,omitempty"`
	BusinessName string `json:"business_name,omitempty"`
}

// NumberCheckResult is returned by CheckNumbers.
type NumberCheckResult struct {
	Number string `json:"number"`
	Exists bool   `json:"exists"`
	JID    string `json:"jid,omitempty"`
}
