package price

import (
	cnt "collection-center/contract/constant"
	"collection-center/internal/logger"
	"collection-center/library/redis"
	"fmt"
	"golang.org/x/xerrors"
	"time"
)

//func init() {
//	fmt.Println(" 测试price文件的init函数")
//}

func SyncEthPriceFeed() error {
	onChainPrice, err := EthPerUSDT()
	if err != nil {
		errNotice := fmt.Sprintf("Sync ETH price error:%s\n", err)
		logger.Error(errNotice)

		return xerrors.New(errNotice)
	}

	_, err = redis.InsertChainData(cnt.EAPU, onChainPrice.String())
	if err != nil {
		errNotice := fmt.Sprintf("Insert ETH price error:%s\n", err)
		logger.Error(errNotice)

		return xerrors.New(errNotice)
	}

	logger.Infof("Insert ETH price to redis successful")

	return nil
}

func SyncBtcPriceFeed() error {
	onChainPrice, err := BtcPerUSDT()
	if err != nil {
		errNotice := fmt.Sprintf("Sync BTC price error:%s\n", err)
		logger.Error(errNotice)

		return xerrors.New(errNotice)
	}

	_, err = redis.InsertChainData(cnt.BAPU, onChainPrice.String())
	if err != nil {
		errNotice := fmt.Sprintf("Insert BTC price error:%s\n", err)
		logger.Error(errNotice)

		return xerrors.New(errNotice)
	}

	logger.Infof("Insert BTC price to redis successful")

	return nil
}

func SyncGasPriceFeed() error {
	onChainPrice, err := GasPriceOnChain()
	if err != nil {
		errNotice := fmt.Sprintf("Sync GAS price error:%s\n", err)
		logger.Error(errNotice)

		return xerrors.New(errNotice)
	}

	_, err = redis.InsertChainData(cnt.PPG, onChainPrice.String())
	if err != nil {
		errNotice := fmt.Sprintf("Insert GAS price error:%s\n", err)
		logger.Error(errNotice)

		return xerrors.New(errNotice)
	}

	logger.Infof("Insert GAS price to redis successful")

	return nil
}

//func inert(field string, price string) (bool, error) {
//	client := redis.Client()
//	status, err := client.Ping().Result()
//	if err != nil {
//		return false, err
//	}
//
//	if status != "PONG" {
//		return false, xerrors.New("Redis链接失败")
//	}
//
//	setCMD := client.Set(field, price, cnt.PRICE_EXPIRED_REDIS)
//
//	setStatus, err := setCMD.Result()
//	if err != nil {
//		return false, err
//	}
//
//	if setStatus != "OK" {
//		return false, xerrors.New(fmt.Sprintf("Insert %s price failed", field))
//	}
//
//	return true, nil
//}

func SyncPrice(period time.Duration) {
	for {
		err := SyncEthPriceFeed()
		if err != nil {
			logger.Error(err)
		}
		err1 := SyncBtcPriceFeed()
		if err1 != nil {
			logger.Error(err1)
		}
		err2 := SyncGasPriceFeed()
		if err2 != nil {
			logger.Error(err2)
		}

		time.Sleep(period)
	}
}
