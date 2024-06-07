package utils

import "encoding/hex"

// HexadecimalToString 16进制转字符串
func HexadecimalToString(str string) (string, error) {
	decodeString, err := hex.DecodeString(str[2:])
	if err != nil {
		return "", err
	}
	return string(decodeString), nil
}
