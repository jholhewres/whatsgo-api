package api

import "time"

// --- Generic ---

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// --- Instance ---

type CreateInstanceRequest struct {
	Name string `json:"name"`
}

type CreateInstanceResponse struct {
	Instance InstanceInfo `json:"instance"`
}

type InstanceInfo struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Token        string    `json:"token,omitempty"`
	Status       string    `json:"status"`
	Phone        string    `json:"phone,omitempty"`
	BusinessName string    `json:"business_name,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type ConnectRequest struct {
	PhoneNumber string `json:"phone_number,omitempty"`
}

type ConnectResponse struct {
	Status      string `json:"status"`
	QRCode      string `json:"qr_code,omitempty"`
	PairingCode string `json:"pairing_code,omitempty"`
}

type StatusResponse struct {
	Instance InstanceInfo `json:"instance"`
}

// --- Messages ---

type SendTextRequest struct {
	Number  string `json:"number"`
	Text    string `json:"text"`
	ReplyTo string `json:"reply_to,omitempty"`
}

type SendMediaRequest struct {
	Number    string `json:"number"`
	MediaType string `json:"media_type"`
	Media     string `json:"media"`
	Caption   string `json:"caption,omitempty"`
	Filename  string `json:"filename,omitempty"`
	MimeType  string `json:"mime_type,omitempty"`
}

type SendLocationRequest struct {
	Number    string  `json:"number"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Name      string  `json:"name,omitempty"`
	Address   string  `json:"address,omitempty"`
}

type SendContactRequest struct {
	Number      string        `json:"number"`
	ContactName string        `json:"contact_name"`
	Phones      []ContactPhone `json:"phones"`
}

type ContactPhone struct {
	Number string `json:"number"`
	Type   string `json:"type,omitempty"`
}

type SendReactionRequest struct {
	Number    string `json:"number"`
	MessageID string `json:"message_id"`
	Emoji     string `json:"emoji"`
}

type SendStickerRequest struct {
	Number   string `json:"number"`
	Sticker  string `json:"sticker"`
	MimeType string `json:"mime_type,omitempty"`
}

type SendResponse struct {
	MessageID string    `json:"message_id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// --- Chat ---

type CheckNumberRequest struct {
	Numbers []string `json:"numbers"`
}

type CheckNumberResponse struct {
	Results []NumberCheck `json:"results"`
}

type NumberCheck struct {
	Number     string `json:"number"`
	Exists     bool   `json:"exists"`
	JID        string `json:"jid,omitempty"`
}

type MarkReadRequest struct {
	ChatJID    string   `json:"chat_jid"`
	MessageIDs []string `json:"message_ids"`
}

type DeleteMessageRequest struct {
	ChatJID   string `json:"chat_jid"`
	MessageID string `json:"message_id"`
	ForMe     bool   `json:"for_me,omitempty"`
}

type EditMessageRequest struct {
	ChatJID   string `json:"chat_jid"`
	MessageID string `json:"message_id"`
	Text      string `json:"text"`
}

type SendPresenceRequest struct {
	ChatJID  string `json:"chat_jid"`
	Presence string `json:"presence"` // composing, paused, recording
}

type BlockRequest struct {
	JID    string `json:"jid"`
	Action string `json:"action"` // block, unblock
}

type ProfileResponse struct {
	JID          string `json:"jid"`
	Name         string `json:"name,omitempty"`
	Status       string `json:"status,omitempty"`
	PictureURL   string `json:"picture_url,omitempty"`
	BusinessName string `json:"business_name,omitempty"`
}

// --- Group ---

type CreateGroupRequest struct {
	Name         string   `json:"name"`
	Participants []string `json:"participants"`
}

type CreateGroupResponse struct {
	JID  string `json:"jid"`
	Name string `json:"name"`
}

type GroupInfo struct {
	JID           string            `json:"jid"`
	Name          string            `json:"name"`
	Topic         string            `json:"topic"`
	OwnerJID      string            `json:"owner_jid"`
	CreatedAt     time.Time         `json:"created_at"`
	Participants  []GroupParticipant `json:"participants"`
	Locked        bool              `json:"locked"`
	Announce      bool              `json:"announce"`
}

type GroupParticipant struct {
	JID     string `json:"jid"`
	IsAdmin bool   `json:"is_admin"`
	IsSuperAdmin bool `json:"is_super_admin"`
}

type UpdateGroupNameRequest struct {
	Name string `json:"name"`
}

type UpdateGroupDescriptionRequest struct {
	Description string `json:"description"`
}

type UpdateGroupSettingsRequest struct {
	Locked   *bool `json:"locked,omitempty"`
	Announce *bool `json:"announce,omitempty"`
}

type ManageParticipantsRequest struct {
	Action       string   `json:"action"` // add, remove, promote, demote
	Participants []string `json:"participants"`
}

// --- Webhook ---

type SetWebhookRequest struct {
	URL     string            `json:"url"`
	Events  []string          `json:"events"`
	Headers map[string]string `json:"headers,omitempty"`
	Enabled bool              `json:"enabled"`
}

// --- Settings ---

type UpdateSettingsRequest struct {
	RejectCall        *bool   `json:"reject_call,omitempty"`
	RejectCallMessage *string `json:"reject_call_message,omitempty"`
	GroupsIgnore      *bool   `json:"groups_ignore,omitempty"`
	AlwaysOnline      *bool   `json:"always_online,omitempty"`
	ReadMessages      *bool   `json:"read_messages,omitempty"`
	ReadReceipts      *bool   `json:"read_receipts,omitempty"`
	WebhookBase64     *bool   `json:"webhook_base64,omitempty"`
}
