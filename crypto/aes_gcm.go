package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// EncryptAESGCM encrypts the plaintext using AES-GCM with the provided key.
// The function returns the ciphertext, which includes the nonce.
// The nonce is randomly generated for each encryption and is used for ensuring uniqueness.
// AES-GCM provides both confidentiality and integrity.
func EncryptAESGCM(key, plaintext []byte) ([]byte, error) {
	// Create a new AES cipher with the given key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create a new GCM (Galois/Counter Mode) with the AES block cipher
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate a random nonce
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt the plaintext using the GCM, and prepend the nonce to the ciphertext
	ciphertext := aesgcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// DecryptAESGCM decrypts the ciphertext using AES-GCM with the provided key.
// The function expects the input ciphertext to include the nonce.
// It returns the decrypted plaintext or an error if the decryption fails.
// AES-GCM provides both confidentiality and integrity.
func DecryptAESGCM(key, ciphertext []byte) ([]byte, error) {
	// Create a new AES cipher with the given key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create a new GCM (Galois/Counter Mode) with the AES block cipher
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Check if the ciphertext is long enough to contain the nonce
	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Separate the nonce from the actual ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the ciphertext using the GCM
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
