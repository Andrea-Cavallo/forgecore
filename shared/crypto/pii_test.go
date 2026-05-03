package crypto

import "testing"

func TestPIIEncryptorCheckedRejectsShortKey(t *testing.T) {
	_, err := NewPIIEncryptorChecked([]byte("short"), []byte("pepper-sufficient"))
	if err == nil {
		t.Fatal("atteso errore per chiave corta")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	pepper := []byte("pepper-sufficient")
	enc, err := NewPIIEncryptorChecked(key, pepper)
	if err != nil {
		t.Fatalf("encryptor: %v", err)
	}
	ciphertext, err := enc.Encrypt("andrea@example.com")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	plaintext, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if plaintext != "andrea@example.com" {
		t.Fatalf("plaintext inatteso: %s", plaintext)
	}
}
