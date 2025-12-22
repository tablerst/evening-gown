package security

import "testing"

func TestHashPassword_WeakPasswordRejected(t *testing.T) {
	_, err := HashPassword("short")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestHashPassword_And_CheckPassword(t *testing.T) {
	h, err := HashPassword("this-is-a-strong-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if h == "" {
		t.Fatalf("expected non-empty hash")
	}
	if !CheckPassword(h, "this-is-a-strong-password") {
		t.Fatalf("expected password to match")
	}
	if CheckPassword(h, "wrong-password") {
		t.Fatalf("expected password not to match")
	}
}
