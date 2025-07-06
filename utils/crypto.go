package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"os"
)

func getEncryptionKey() []byte {
	key := os.Getenv("ENCRYPTION_KEY")
	if key == "" {
		key = "mi-clave-super-secreta-32-chars!!" // 32 chars para AES-256
	}
	return []byte(key)[:32] // Asegurar 32 bytes
}

func EncryptField(plaintext string) string {
	if plaintext == "" {
		return ""
	}

	key := getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return plaintext // En caso de error, devolver sin encriptar
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return plaintext
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return plaintext
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

func DecryptField(ciphertext string) string {
	if ciphertext == "" {
		return ""
	}

	key := getEncryptionKey()
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return ciphertext // En caso de error, devolver como está
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return ciphertext
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return ciphertext
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return ciphertext
	}

	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return ciphertext
	}

	return string(plaintext)
}
