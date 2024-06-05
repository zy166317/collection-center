package utils

import (
	"golang.org/x/xerrors"
	"math/big"
	"math/rand"
)

func RandomEthRpcUrl(urls []string) (string, int, error) {
	urlsLen := len(urls)
	if urlsLen == 0 {
		return "", 0, xerrors.New("ETH urls not set up")
	}
	//rand.New(rand.NewSource(time.Now().UnixNano())) // init-ed at main.go
	// 生成一个范围随机整数
	randomInt := rand.Intn(urlsLen)
	return urls[randomInt], randomInt, nil
}

func MatchNetwork(id *big.Int) string {
	var name string

	switch id.Int64() {
	case 1:
		name = "Mainnet"
		break
	case 5:
		name = "Goerli"
		break
	default:
		name = "Null"
		break
	}

	return name
}
