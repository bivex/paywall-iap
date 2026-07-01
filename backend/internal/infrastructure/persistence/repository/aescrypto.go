package repository

// aescrypto provides AES-256-GCM encrypt/decrypt for credential fields.
// Key must be exactly 32 bytes, injected via APP_CREDENTIALS_KEY env var.
// Empty key → no-op (plaintext stored), useful in dev/test.

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// encryptField encrypts plaintext with AES-256-GCM and returns base64(nonce+ciphertext).
// Returns "" if plaintext is "".
func encryptField(key []byte, plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	if len(key) == 0 {
		return plaintext, nil // dev mode: no encryption
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("aes: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("nonce: %w", err)
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// decryptField decrypts a base64(nonce+ciphertext) produced by encryptField.
// Returns "" if ciphertext is "".
func decryptField(key []byte, ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	if len(key) == 0 {
		return ciphertext, nil // dev mode: no encryption
	}
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("base64: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("aes: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("gcm: %w", err)
	}
	ns := gcm.NonceSize()
	if len(data) < ns {
		return "", errors.New("ciphertext too short")
	}
	plain, err := gcm.Open(nil, data[:ns], data[ns:], nil)
	if err != nil {
		return "", fmt.Errorf("gcm open: %w", err)
	}
	return string(plain), nil
}
