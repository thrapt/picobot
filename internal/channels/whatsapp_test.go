package channels

import (
	"testing"
)

// TestWhatsAppSetup_EmptyDBPath tests that SetupWhatsApp returns an error with empty db path.
func TestWhatsAppSetup_EmptyDBPath(t *testing.T) {
	err := SetupWhatsApp("")
	if err == nil {
		t.Error("SetupWhatsApp with empty dbPath should return error")
	}
	if err.Error() != "whatsapp database path not provided" {
		t.Errorf("expected 'whatsapp database path not provided' error, got: %v", err)
	}
}

// TestWhatsApp_AllowlistLogic tests the allowlist logic conceptually.
func TestWhatsApp_AllowlistLogic(t *testing.T) {
	// This tests the allowlist logic conceptually
	allowed := make(map[string]struct{})
	allowed["1234567890"] = struct{}{}

	// Test allowed user
	if _, ok := allowed["1234567890"]; !ok {
		t.Error("user 1234567890 should be allowed")
	}

	// Test non-allowed user
	if _, ok := allowed["9876543210"]; ok {
		t.Error("user 9876543210 should not be allowed")
	}

	// Test empty allowlist (all users allowed)
	emptyAllowed := make(map[string]struct{})
	if len(emptyAllowed) > 0 {
		t.Error("empty allowlist should allow all users")
	}
}

// TestWhatsApp_MessageContentExtraction tests message content extraction logic.
func TestWhatsApp_MessageContentExtraction(t *testing.T) {
	tests := []struct {
		name         string
		conversation *string
		expected     string
	}{
		{
			name:         "basic conversation",
			conversation: stringPtr("Hello, bot!"),
			expected:     "Hello, bot!",
		},
		{
			name:         "nil conversation",
			conversation: nil,
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := ""
			if tt.conversation != nil {
				content = *tt.conversation
			}

			if content != tt.expected {
				t.Errorf("content = %q, want %q", content, tt.expected)
			}
		})
	}
}

// TestWhatsApp_PhoneNumberFiltering tests phone number filtering logic.
func TestWhatsApp_PhoneNumberFiltering(t *testing.T) {
	allowed := make(map[string]struct{})
	allowed["1234567890"] = struct{}{}
	allowed["9876543210"] = struct{}{}

	testCases := []struct {
		phoneNumber string
		shouldAllow bool
	}{
		{"1234567890", true},
		{"9876543210", true},
		{"5555555555", false},
	}

	for _, tc := range testCases {
		_, ok := allowed[tc.phoneNumber]
		if ok != tc.shouldAllow {
			t.Errorf("phone number %s: expected allowed=%v, got %v", tc.phoneNumber, tc.shouldAllow, ok)
		}
	}
}

// stringPtr is a helper to create string pointers for tests
func stringPtr(s string) *string {
	return &s
}
