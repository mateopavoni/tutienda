package accounts

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestFailInvalidCredentials pins the contract both login endpoints promise: an unregistered email and a
// wrong password get the exact same status and the exact same message, and differ only in clear_email —
// the flag the login form uses to decide whether blanking the email field could possibly help. If the two
// messages ever drift apart, the UI starts telling users which half they got wrong; this test is what
// keeps that from happening quietly.
func TestFailInvalidCredentials(t *testing.T) {
	respond := func(err error) (int, loginFailure) {
		w := httptest.NewRecorder()
		failInvalidCredentials(w, err)
		var got loginFailure
		if decErr := json.NewDecoder(w.Body).Decode(&got); decErr != nil {
			t.Fatalf("decode body: %v", decErr)
		}
		return w.Code, got
	}
	badPassStatus, badPass := respond(ErrInvalidCredentials)
	unknownStatus, unknown := respond(ErrUnknownEmail)

	if badPassStatus != http.StatusUnauthorized || unknownStatus != http.StatusUnauthorized {
		t.Errorf("both must be 401, got %d and %d", badPassStatus, unknownStatus)
	}
	if badPass.Error != unknown.Error {
		t.Errorf("messages must be identical, got %q and %q", badPass.Error, unknown.Error)
	}
	if badPass.ClearEmail {
		t.Error("a wrong password must keep the email the user typed")
	}
	if !unknown.ClearEmail {
		t.Error("an email with no account behind it must be cleared")
	}
	// The wrapping is what keeps every existing errors.Is(err, ErrInvalidCredentials) check — including
	// the one that routes to this very response — working for the unknown-email case.
	if !errors.Is(ErrUnknownEmail, ErrInvalidCredentials) {
		t.Error("ErrUnknownEmail must still be an ErrInvalidCredentials")
	}
}
