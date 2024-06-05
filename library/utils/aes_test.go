package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"

	"golang.org/x/crypto/pbkdf2"
)

func TestGetDigest(t *testing.T) {
	t.Log(getDecimalsInt(2))
}

func TestAes(t *testing.T) {
	password := "12345678"

	plainText := "1cVrCXw7WpgUthSyn48pyNDvEYCj62jYnNojVdR67GkHrfYJSMHfK"
	// rK/QlszkBAVniHkRR+Rq0eFXCmlTNWquqDfy8vmj+DnzFDSA7WmGvlRqSP5AIanTSdNfHr64DPUKTv1NUfyzmw==

	//plainText := "2f219ebf353f8c3f5c3cd691d03b92356b9c1bb0f29f80a4fa759d8af2d26dd2"
	// fuvV73gXBIJ0x4YPA4ghO3mZ0zbEIgAfuBmoEHSGLJAgiYhR5SgffVpAtvEzO4It7N07ZwqNmVRIiqO/1yna5hVdviN5FLNcuQ5QktUqUSE=
	dk := pbkdf2.Key([]byte(password), []byte(salt), 1024, 32, sha256.New)

	fmt.Println("", dk)
	str4 := hex.EncodeToString(dk)
	fmt.Println("", str4)

	fmt.Println(len(str4))

	encryptedData, err := aesEncrypt(plainText, dk)
	if err != nil {
		t.Fatal(err)
	}

	encryptedString := base64.RawURLEncoding.EncodeToString(encryptedData)
	fmt.Println(encryptedString)

	encryptedDataB, _ := base64.RawURLEncoding.DecodeString(encryptedString)
	decryptedText, err := aesDecrypt(encryptedDataB, dk)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(plainText)
	fmt.Println(string(decryptedText))
	fmt.Println(string(decryptedText) == plainText)
}

func TestEncrypt(t *testing.T) {
	// password := "888899998888"
	SetPasswd("888899998888")
	plainText := "1q2w3e4r5t6y7u8i"
	cipherText, err := Encrypt(plainText)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("============")
	t.Log(cipherText)
	t.Log("============")

	p, err := Decrypt(cipherText)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("============")
	t.Log(p)
	t.Log("============")
}
