package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
)

const (
	AES256KeySize = 32
	MinPepperSize = 16
)

// PIIEncryptor encrypts and decrypts personally identifiable information using AES-256-GCM.
// It also provides HMAC-SHA256 deterministic hashing for DB index lookups.
type PIIEncryptor struct {
	key    []byte
	pepper []byte
}

// NewPIIEncryptor creates a new encryptor with a 32-byte AES-256 key and a pepper for HMAC hashing.
func NewPIIEncryptor(key, pepper []byte) *PIIEncryptor {
	return &PIIEncryptor{key: key, pepper: pepper}
}

func NewPIIEncryptorChecked(key, pepper []byte) (*PIIEncryptor, error) {
	if len(key) != AES256KeySize {
		return nil, fmt.Errorf("chiave AES-256 non valida: %d byte", len(key))
	}
	if len(pepper) < MinPepperSize {
		return nil, fmt.Errorf("pepper HMAC troppo corto: %d byte", len(pepper))
	}
	return NewPIIEncryptor(key, pepper), nil
}

// Hash returns a deterministic HMAC-SHA256 hex hash using the encryptor pepper.
// Use this for DB indexes on encrypted PII fields (e.g., email lookup).
func (e *PIIEncryptor) Hash(value string) string {
	mac := hmac.New(sha256.New, e.pepper)
	mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

// Encrypt encrypts plaintext and returns a base64-encoded ciphertext.
func (e *PIIEncryptor) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("new gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("rand nonce: %w", err)
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64-encoded ciphertext back to plaintext.
func (e *PIIEncryptor) Decrypt(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("new gcm: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("gcm open: %w", err)
	}
	return string(plain), nil
}
