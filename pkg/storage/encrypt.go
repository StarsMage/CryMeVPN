package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"errors"
)

var nothing = []byte{}

func Encrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key) 

	if err != nil {
		return nothing,err
	}

	gcm, err := cipher.NewGCM(block)

	if err != nil {
		return nothing,err
	}

	nonce := make([]byte, gcm.NonceSize())

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nothing,err
	}

	return gcm.Seal(nonce,nonce,plaintext,nil), nil
}

func Decrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return nothing, err
	}

	gcm, err := cipher.NewGCM(block)

	if err != nil {
		return nothing, err
	}

	nonceSize := gcm.NonceSize()

	if len(plaintext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce := plaintext[:nonceSize]
	encryptedData := plaintext[nonceSize:]
	return gcm.Open(nil,nonce, encryptedData,nil)
}