package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

var CryptKey = os.Getenv("SONORA_ENCRYPT_KEY")

// Encrypt returns a single blob: nonce || ciphertext
func Encrypt(key, plaintext string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// ciphertext = plaintext + tag
	ciphertext := aesgcm.Seal(nil, nonce, []byte(plaintext), nil)

	// store as nonce||ciphertext
	return base64.StdEncoding.EncodeToString(append(nonce, ciphertext...)), nil
}

// Decrypt expects blob = nonce || ciphertext
func Decrypt(key, blob string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	blobBytes, err := base64.StdEncoding.DecodeString(blob)
	if err != nil {
		return "", err
	}

	nonceSize := aesgcm.NonceSize()
	if len(blobBytes) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := blobBytes[:nonceSize], blobBytes[nonceSize:]
	if plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil); err != nil {
		return "", err
	} else {
		return string(plaintext), nil
	}
}
