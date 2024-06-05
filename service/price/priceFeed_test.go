package price

import (
	cnt "collection-center/contract/constant"
	"collection-center/library/redis"
	"fmt"
	"testing"
	"time"
)

func TestPricefeed(t *testing.T) {
	client, err := redis.SetRedis(&redis.RedisConfig{
		Addr:         "192.168.8.63:6379",
		Auth:         "orca_redis",
		DialTimeout:  0,
		ReadTimeout:  0,
		WriteTimeout: 0,
	})
	if err != nil {
		t.Errorf("Connect readis error:%v\n", err)
	}

	SyncEthPriceFeed()
	ret := client.Get(cnt.EAPU)
	num, _ := ret.Int()
	fmt.Printf("EAPU:%d\n", num)

	SyncBtcPriceFeed()
	ret2 := client.Get(cnt.BAPU)
	num2, _ := ret2.Int()
	fmt.Printf("BAPU:%d\n", num2)

	SyncGasPriceFeed()
	ret3 := client.Get(cnt.PPG)
	num3, _ := ret3.Int()
	fmt.Printf("PPG:%d\n", num3)
}

// 测试高并发的时候，获取价格接口
func TestLoop(t *testing.T) {
	for i := 0; i < 1000; i++ {
		go func() {
			SyncEthPriceFeed()
		}()
	}
	time.Sleep(100 * time.Second)
}
