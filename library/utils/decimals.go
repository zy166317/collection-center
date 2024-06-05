package utils

import (
	"errors"
	"fmt"
	"golang.org/x/xerrors"
	"math/big"
	"strings"
)

// getDecimalsInt 获取精度 10^precision
func getDecimalsInt(precision int) *big.Int {
	base := new(big.Int).SetInt64(10)
	exponent := new(big.Int).SetInt64(int64(precision))
	modulus := new(big.Int).SetInt64(0)
	return new(big.Int).Exp(base, exponent, modulus)
}

const (
	CoinEth  = "ETH"
	CoinBtc  = "BTC"
	CoinUsdt = "USDT"
)

func GetDecimals(coinType string) (*big.Int, *big.Float) {
	if coinType == CoinUsdt {
		res := getDecimalsInt(6)
		return res, new(big.Float).SetInt(res)
	} else if coinType == CoinEth {
		res := getDecimalsInt(18)
		return res, new(big.Float).SetInt(res)
	} else if coinType == CoinBtc {
		res := getDecimalsInt(8)
		return res, new(big.Float).SetInt(res)
	} else {
		return big.NewInt(0), big.NewFloat(0)
	}
}

func WeiToEth(wei *big.Int) *big.Float {
	_, ethDecimals := GetDecimals(CoinEth)
	return new(big.Float).Quo(new(big.Float).SetInt(wei), ethDecimals)
}

func EthToWei(eth *big.Float) *big.Int {
	_, ethDecimals := GetDecimals(CoinEth)
	weiAmount := new(big.Float).Mul(eth, ethDecimals)
	weiAmountInt, _ := weiAmount.Int(nil)
	return weiAmountInt
}

func BtcToSatoshi(btc *big.Float) (*big.Int, error) {
	_, btcDecimals := GetDecimals(CoinBtc)
	satoshiAmount := new(big.Float).Mul(btc, btcDecimals)
	satoshiAmountInt, _ := satoshiAmount.Int(nil)
	return satoshiAmountInt, nil
}

func SatoshiToBtc(satoshi *big.Int) (*big.Float, error) {
	_, btcDecimals := GetDecimals(CoinBtc)
	btcAmount := new(big.Float).Quo(new(big.Float).SetInt(satoshi), btcDecimals)
	return btcAmount, nil
}

func StrToBigInt(num string) (*big.Int, error) {
	amount, status := new(big.Int).SetString(num, 10)
	if !status {
		return big.NewInt(0), xerrors.New("Convert string number error")
	}

	return amount, nil
}

func StrToBigFloat(num string) (*big.Float, error) {
	floatAmount, status := new(big.Float).SetString(num)
	if !status {
		return big.NewFloat(0), xerrors.New("Convert string number error")
	}

	return floatAmount, nil
}

func AsStringFromFloat(precision int, amount *big.Float) (string, error) {
	fmtString := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(strings.TrimRight(strings.TrimRight(fmt.Sprintf(fmtString, amount), "0"), ".")), nil
}

// MustToFloat 强转
func MustToFloat(amount *big.Float) float64 {
	amountFloat, _ := amount.Float64()
	return amountFloat
}

// 将最小单位的数据转换为可读单位
func DecimalParse(amount interface{}, decimal int) (string, error) {
	var amountTemp *big.Float
	switch amount.(type) {
	case string:
		amount0, status := new(big.Float).SetString(amount.(string))
		if !status {
			return "", errors.New("amount type error")
		}
		amountTemp = amount0
	case int:
		amountTemp = new(big.Float).SetInt64(int64(amount.(int)))
	case int32:
		amountTemp = new(big.Float).SetInt64(int64(amount.(int32)))
	case int64:
		amountTemp = new(big.Float).SetInt64(amount.(int64))
	default:
		return "", errors.New("amount type error")
	}
	base := new(big.Int).SetInt64(10)
	exponent := new(big.Int).SetInt64(int64(decimal))
	modulus := new(big.Int).SetInt64(0)
	decimalBigInt := new(big.Int).Exp(base, exponent, modulus)
	decimalFloat := new(big.Float).SetInt(decimalBigInt)
	amountOut := new(big.Float).Quo(amountTemp, decimalFloat)
	amountStr, err := AsStringFromFloat(decimal, amountOut)
	return amountStr, err
}
