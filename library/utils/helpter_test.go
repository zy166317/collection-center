package utils

import (
	"fmt"
	"testing"
)

func init() {
	//rpc.WaitBlockCltEthMax = 25
	//rpc.WaitBlockCltEthMin = 20
	//btc.WaitBlockCltBtcMax = 1
	//btc.WaitBlockCltBtcMin = 2
}

func TestRangeRandom(t *testing.T) {
	num := RangeRandom(1, 2)
	fmt.Printf("Random number:%d\n", num)
}
