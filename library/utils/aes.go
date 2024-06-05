package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"errors"

	"golang.org/x/crypto/pbkdf2"
)

var (
	initialVector = "1234567890123456"
	salt          = "2e54d4e12c8141fff46a62f0ba29769e623de1e7917eda8e8ee5f8ffb05d5fde"
	passwd        = ""
)

func SetPasswd(passwdStr string) {
	passwd = passwdStr
}
func Encrypt(plainText string) (string, error) {
	key := passwd
	dk := pbkdf2.Key([]byte(key), []byte(salt), 1024, 32, sha256.New)
	encryptedData, err := aesEncrypt(plainText, dk)
	if err != nil {
		return "", err
	}
	encryptedString := base64.RawURLEncoding.EncodeToString(encryptedData)
	return encryptedString, nil
}

func Decrypt(cipherText string) (string, error) {
	key := passwd
	encryptedDataB, err := base64.RawURLEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}
	dk := pbkdf2.Key([]byte(key), []byte(salt), 1024, 32, sha256.New)
	decryptedText, err := aesDecrypt(encryptedDataB, dk)
	if err != nil {
		return "", err
	}
	return string(decryptedText), nil
}

func aesEncrypt(src string, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if src == "" {
		return nil, errors.New("plain content empty")
	}
	ecb := cipher.NewCBCEncrypter(block, []byte("1234567890123456"))
	content := []byte(src)
	content = pKCS5Padding(content, block.BlockSize())
	crypted := make([]byte, len(content))
	ecb.CryptBlocks(crypted, content)

	return crypted, nil
}

func aesDecrypt(crypt []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(crypt) == 0 {
		return nil, errors.New("plain content empty")
	}
	ecb := cipher.NewCBCDecrypter(block, []byte(initialVector))
	decrypted := make([]byte, len(crypt))
	ecb.CryptBlocks(decrypted, crypt)
	return pKCS5Trimming(decrypted), nil
}

func pKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pKCS5Trimming(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}
