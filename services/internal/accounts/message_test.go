package accounts

import (
	"errors"
	"strings"
	"testing"
)

// TestNewMessage pins the contact-form validation: name/email/body are required, email must look like
// one, and everything is trimmed/capped before it would ever reach the DB.
func TestNewMessage(t *testing.T) {
	m, err := newMessage("store1", "  Ada  ", "  ADA@Example.com  ", "  hi  "+strings.Repeat("x", messageBodyMax))
	if err != nil {
		t.Fatalf("valid message rejected: %v", err)
	}
	if m.TenantID != "store1" || m.Name != "Ada" || m.Email != "ada@example.com" {
		t.Errorf("not trimmed/normalized: %+v", m)
	}
	if len([]rune(m.Body)) != messageBodyMax {
		t.Errorf("body not capped to %d runes: %d", messageBodyMax, len([]rune(m.Body)))
	}
}

func TestNewMessageInvalid(t *testing.T) {
	cases := map[string]struct{ name, email, body string }{
		"blank name":  {"", "a@b.com", "hi"},
		"blank email": {"Ada", "", "hi"},
		"email no at": {"Ada", "not-an-email", "hi"},
		"blank body":  {"Ada", "a@b.com", "   "},
	}
	for label, c := range cases {
		if _, err := newMessage("store1", c.name, c.email, c.body); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("%s: want ErrInvalidInput, got %v", label, err)
		}
	}
}
