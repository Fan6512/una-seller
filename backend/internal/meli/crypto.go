package meli

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"strings"
)

const tokenCipherVersion byte = 1

// TokenCipher protects Mercado Libre OAuth tokens in the private database.
// The key is 32 random bytes, provided as base64 from the SERVER environment.
type TokenCipher struct {
	aead cipher.AEAD
}

func NewTokenCipher(encodedKey string) (*TokenCipher, error) {
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encodedKey))
	if err != nil || len(key) != 32 {
		return nil, errors.New("MELI_TOKEN_ENCRYPTION_KEY must be base64 of 32 random bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.New("invalid token encryption key")
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.New("cannot initialize token encryption")
	}
	return &TokenCipher{aead: aead}, nil
}

// Seal includes a random nonce and format version. AAD binds each token to its owner,
// platform seller, and purpose, preventing ciphertext movement between DB rows.
func (c *TokenCipher) Seal(value string, aad []byte) ([]byte, error) {
	if c == nil || c.aead == nil || value == "" {
		return nil, errors.New("token encryption unavailable")
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errors.New("cannot generate token nonce")
	}
	out := make([]byte, 1+len(nonce))
	out[0] = tokenCipherVersion
	copy(out[1:], nonce)
	out = c.aead.Seal(out, nonce, []byte(value), aad)
	return out, nil
}

func (c *TokenCipher) Open(blob, aad []byte) (string, error) {
	if c == nil || c.aead == nil || len(blob) < 1+c.aead.NonceSize()+c.aead.Overhead() ||
		blob[0] != tokenCipherVersion {
		return "", errors.New("invalid encrypted token")
	}
	nonceEnd := 1 + c.aead.NonceSize()
	plain, err := c.aead.Open(nil, blob[1:nonceEnd], blob[nonceEnd:], aad)
	if err != nil {
		return "", errors.New("encrypted token authentication failed")
	}
	return string(plain), nil
}
