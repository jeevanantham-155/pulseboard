package auth

import "testing"

func TestValidateSignup(t *testing.T) {
	if err := validateSignup("A", "bad", "short"); err == nil {
		t.Fatal("expected invalid signup details")
	}
	if err := validateSignup("Ada Lovelace", "ada@example.com", "correct horse battery staple"); err != nil {
		t.Fatalf("expected valid signup, got %v", err)
	}
}
