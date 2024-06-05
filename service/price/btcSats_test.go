package price

import (
	"collection-center/internal/btc"
	"testing"
)

func TestSyncSats(t *testing.T) {
	//sat, err  := getSat()
	//if err != nil {
	//	t.Errorf("Get sat error:%v\n", err)
	//}
	//
	//fmt.Printf("SAT:%d\n", sat)
	btc.InitBtcd(true)
	testSAT()
}
