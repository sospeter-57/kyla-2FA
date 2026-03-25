package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
)

var ErrInvalidPIN = errors.New("invalid PIN or encrypted file corrupted")

func storagePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(configDir, "kyla-2FA", "vault.dat")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return "", err
	}
	return path, nil
}

func deriveKey(pin string) []byte {
	h := sha256.Sum256([]byte(pin))
	return h[:]
}

func encryptData(key []byte, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	cipherText := gcm.Seal(nil, nonce, plaintext, nil)
	return append(nonce, cipherText...), nil
}

func decryptData(key []byte, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrInvalidPIN
	}
	nonce, cipherData := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return nil, ErrInvalidPIN
	}
	return plaintext, nil
}

func saveVault(pin string) error {
	path, err := storagePath()
	if err != nil {
		return err
	}
	payload, err := json.Marshal(vault)
	if err != nil {
		return err
	}
	encrypted, err := encryptData(deriveKey(pin), payload)
	if err != nil {
		return err
	}
	return os.WriteFile(path, encrypted, 0600)
}

func loadVault(pin string) (Vault, error) {
	path, err := storagePath()
	if err != nil {
		return Vault{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Vault{}, err
	}
	plaintext, err := decryptData(deriveKey(pin), data)
	if err != nil {
		return Vault{}, err
	}
	var v Vault
	if err := json.Unmarshal(plaintext, &v); err != nil {
		return Vault{}, err
	}
	return v, nil
}

func backupToFile(dst string) error {
	payload, err := json.MarshalIndent(vault, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dst, payload, 0600)
}

func restoreFromFile(src string) error {
	payload, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	var v Vault
	if err := json.Unmarshal(payload, &v); err != nil {
		return err
	}
	vault = v
	return saveVault(activePIN)
}
