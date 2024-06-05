package price

import (
	"collection-center/library/redis"
	"fmt"
	"testing"
)

func init() {
	redis.SetRedis(&redis.RedisConfig{
		Addr:         "192.168.8.63:6379",
		Auth:         "orca_redis",
		DialTimeout:  0,
		ReadTimeout:  0,
		WriteTimeout: 0,
	})
}

func TestMultiPrice(t *testing.T) {
	bFloatPrice, bIntPrice, err := MultiTypeChainPrice()
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("Big Float price:%+v\n", bFloatPrice)
	fmt.Printf("Big Int price:%+v\n", bIntPrice)

}
