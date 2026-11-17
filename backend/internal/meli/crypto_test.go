package meli

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestTokenCipherRoundTripAndIsolation(t *testing.T) {
	key := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x42}, 32))
	c, err := NewTokenCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	aad := []byte("owner-A|seller-1|refresh")
	sealed, err := c.Seal("synthetic-refresh-token", aad)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(sealed, []byte("synthetic-refresh-token")) {
		t.Fatal("plaintext token appeared in ciphertext")
	}
	got, err := c.Open(sealed, aad)
	if err != nil || got != "synthetic-refresh-token" {
		t.Fatalf("roundtrip failed: %v", err)
	}
	if _, err := c.Open(sealed, []byte("owner-B|seller-1|refresh")); err == nil {
		t.Fatal("wrong owner must not decrypt a token")
	}
	if _, err := c.Open(sealed, []byte("owner-A|seller-1|access")); err == nil {
		t.Fatal("wrong token purpose must not decrypt a token")
	}
	sealed[len(sealed)-1] ^= 1
	if _, err := c.Open(sealed, aad); err == nil {
		t.Fatal("tampered ciphertext must be rejected")
	}
}

func TestTokenCipherInvalidInputs(t *testing.T) {
	if _, err := NewTokenCipher(""); err == nil {
		t.Fatal("missing key was accepted")
	}
	if _, err := NewTokenCipher(base64.StdEncoding.EncodeToString(make([]byte, 16))); err == nil {
		t.Fatal("short key was accepted")
	}
	key := base64.StdEncoding.EncodeToString(make([]byte, 32))
	c, err := NewTokenCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Seal("", nil); err == nil {
		t.Fatal("empty token was accepted")
	}
	if _, err := c.Open([]byte{1}, nil); err == nil {
		t.Fatal("short ciphertext was accepted")
	}
}
