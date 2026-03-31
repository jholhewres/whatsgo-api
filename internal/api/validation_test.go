package api

import "testing"

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"5511999999999", true},    // BR mobile
		{"14155552671", true},      // US number
		{"+5511999999999", true},   // with + prefix
		{"55 11 99999-9999", true}, // with spaces/dashes
		{"123456", false},          // too short
		{"1234567890123456", false}, // too long (16 digits)
		{"", false},
		{"abc", false},
		{"1234567", true}, // minimum 7 digits
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ValidatePhone(tt.input); got != tt.valid {
				t.Errorf("ValidatePhone(%q) = %v, want %v", tt.input, got, tt.valid)
			}
		})
	}
}

func TestValidateMediaType(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"image", true},
		{"video", true},
		{"audio", true},
		{"document", true},
		{"IMAGE", true},  // case insensitive
		{"Video", true},
		{"sticker", false},
		{"", false},
		{"pdf", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ValidateMediaType(tt.input); got != tt.valid {
				t.Errorf("ValidateMediaType(%q) = %v, want %v", tt.input, got, tt.valid)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	if MaxTextLength != 65536 {
		t.Fatalf("expected MaxTextLength 65536, got %d", MaxTextLength)
	}
	if MaxCaptionLength != 1024 {
		t.Fatalf("expected MaxCaptionLength 1024, got %d", MaxCaptionLength)
	}
	if MaxGroupName != 100 {
		t.Fatalf("expected MaxGroupName 100, got %d", MaxGroupName)
	}
	if MaxParticipants != 256 {
		t.Fatalf("expected MaxParticipants 256, got %d", MaxParticipants)
	}
}
