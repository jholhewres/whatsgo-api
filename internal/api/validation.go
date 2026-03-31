package api

import (
	"regexp"
	"strings"
)

// phonePattern matches international phone numbers: country code + number (digits only, 7-15 digits).
var phonePattern = regexp.MustCompile(`^\d{7,15}$`)

// ValidatePhone checks if a phone number is in a valid format (digits only, 7-15 chars).
func ValidatePhone(number string) bool {
	cleaned := strings.ReplaceAll(number, "+", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	return phonePattern.MatchString(cleaned)
}

const (
	MaxTextLength    = 65536 // 64KB max text message
	MaxCaptionLength = 1024  // WhatsApp caption limit
	MaxGroupName     = 100   // WhatsApp group name limit
	MaxParticipants  = 256   // WhatsApp group participant limit
)

var validMediaTypes = map[string]bool{
	"image":    true,
	"video":    true,
	"audio":    true,
	"document": true,
}

// ValidateMediaType checks if the media type is supported.
func ValidateMediaType(mediaType string) bool {
	return validMediaTypes[strings.ToLower(mediaType)]
}

var validPresenceTypes = map[string]bool{
	"composing": true,
	"paused":    true,
	"recording": true,
}

var validParticipantActions = map[string]bool{
	"add":     true,
	"remove":  true,
	"promote": true,
	"demote":  true,
}

var validBlockActions = map[string]bool{
	"block":   true,
	"unblock": true,
}
