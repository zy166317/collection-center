package utils

import (
	"fmt"
	"math/big"
)

// HexadecimalToString 16进制转字符串
func HexadecimalToString(str string) string {
	// 去除字符串中的"0x"前缀
	str = str[2:]

	// 将十六进制字符串转换为大整数
	bigInt := new(big.Int)
	bigInt.SetString(str, 16)

	// 将大整数转换为十进制字符串
	decimalStr := bigInt.String()
	return decimalStr
}

func Hex(str string) string {
	// 去除字符串中的"0x"前缀
	str = str[2:]

	// 将十六进制字符串转换为大整数
	bigInt := new(big.Int)
	bigInt.SetString(str, 16)
	hexStrNew := fmt.Sprintf("%x", bigInt)
	return "0x" + hexStrNew
}
