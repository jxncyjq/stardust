package jwt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

func getDerivedKey(password string, salt []byte, count int) []byte {
	key := sha256.Sum256(append([]byte(password), salt...))
	for i := 0; i < count-1; i++ {
		key = sha256.Sum256(key[:])
	}
	return key[:]
}

func Encrypt(password string, obtenationIterations int, plainText string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	key := getDerivedKey(password, salt, obtenationIterations)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	result := append(salt, ciphertext...)
	return base64.StdEncoding.EncodeToString(result), nil
}

func Decrypt(password string, obtenationIterations int, cipherText string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}
	if len(data) < 16 {
		return "", errors.New("ciphertext too short")
	}

	salt := data[:16]
	cipherdata := data[16:]

	key := getDerivedKey(password, salt, obtenationIterations)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(cipherdata) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}

	nonce, cipherdata := cipherdata[:gcm.NonceSize()], cipherdata[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, cipherdata, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// EncryptWithFixedSalt 使用固定salt加密
func EncryptWithFixedSalt(password string, obtenationIterations int, plainText string, fixedSalt string) (string, error) {
	salt := make([]byte, 16)
	copy(salt, fixedSalt)

	key := getDerivedKey(password, salt, obtenationIterations)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptWithFixedSalt 使用固定salt解密
func DecryptWithFixedSalt(password string, obtenationIterations int, cipherText string, fixedSalt string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	salt := make([]byte, 16)
	copy(salt, fixedSalt)

	key := getDerivedKey(password, salt, obtenationIterations)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(data) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}

	nonce, cipherdata := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, cipherdata, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
