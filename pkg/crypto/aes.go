package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
)

// AESCipher wraps AES-256-GCM encryption/decryption.
// Key must be 32 bytes (hex-encoded in env, decoded here).
// Used by UserRepository to encrypt health and injury data at rest.
type AESCipher struct {
	block cipher.Block
}

// NewAESCipher creates a cipher from a 32-byte hex-encoded key (env: AES_KEY).
func NewAESCipher(hexKey string) (*AESCipher, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, err
	}
	if len(key) != 32 {
		return nil, errors.New("AES key must be 32 bytes (64 hex chars)")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &AESCipher{block: block}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM. Returns nonce+ciphertext.
func (c *AESCipher) Encrypt(plaintext []byte) ([]byte, error) {
	gcm, err := cipher.NewGCM(c.block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt reverses Encrypt. Input must be nonce+ciphertext as returned by Encrypt.
func (c *AESCipher) Decrypt(data []byte) ([]byte, error) {
	gcm, err := cipher.NewGCM(c.block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
