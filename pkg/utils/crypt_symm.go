package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

var key_symm = []byte("NAZ1PEJzr0nDuVDF")

func EncryptSymm(plaintext string) (string, error) {
	plainbyte := []byte(plaintext)
	block, err := aes.NewCipher(key_symm)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plainbyte))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], plainbyte)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DecryptSymm(ciphertext string) (string, error) {
	block, err := aes.NewCipher(key_symm)
	if err != nil {
		return "", err
	}

	decodedCiphertext, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	if len(decodedCiphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	iv := decodedCiphertext[:aes.BlockSize]
	decodedCiphertext = decodedCiphertext[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(decodedCiphertext, decodedCiphertext)
	return string(decodedCiphertext), nil
}
