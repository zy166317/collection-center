package price

import (
	"fmt"
	"testing"
)

func TestEthPrice(t *testing.T) {
	a, err := EthPerUSDT()
	t.Error(err)
	b, err := BtcPerUSDT()
	t.Error(err)
	c, err := GasPriceOnChain()
	t.Error(err)
	fmt.Println(a, b, c)
}
