package webhook

import "time"

// Payload is the webhook HTTP POST body.
type Payload struct {
	Event     string    `json:"event"`
	Instance  string    `json:"instance"`
	Data      any       `json:"data"`
	Timestamp time.Time `json:"timestamp"`
	ServerURL string    `json:"server_url"`
}
