package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// Encryptor provides authenticated encryption using AES-GCM
type Encryptor struct {
	key []byte
}

// NewEncryptor creates a new encryptor with the provided key
// Key must be 16, 24, or 32 bytes (AES-128, AES-192, or AES-256)
func NewEncryptor(key []byte) (*Encryptor, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("invalid key size: must be 16, 24, or 32 bytes")
	}
	return &Encryptor{key: key}, nil
}

// Encrypt encrypts the plaintext using AES-GCM and returns base64 encoded result
// The result includes the nonce and ciphertext concatenated together
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	// Create AES cipher block
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and authenticate the plaintext
	plaintextBytes := []byte(plaintext)
	ciphertext := gcm.Seal(nonce, nonce, plaintextBytes, nil)

	// Encode to base64 for safe string representation
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts the base64 encoded ciphertext and verifies authenticity
func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	// Decode from base64
	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Create AES cipher block
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Check minimum length (nonce + some ciphertext)
	nonceSize := gcm.NonceSize()
	if len(ciphertextBytes) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce, ciphertextOnly := ciphertextBytes[:nonceSize], ciphertextBytes[nonceSize:]

	// Decrypt and verify authenticity
	plaintextBytes, err := gcm.Open(nil, nonce, ciphertextOnly, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt or authenticate: %w", err)
	}

	return string(plaintextBytes), nil
}
