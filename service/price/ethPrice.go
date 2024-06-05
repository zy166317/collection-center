package price

import (
	"collection-center/contract/build"
	"collection-center/internal/logger"
	"collection-center/internal/rpc"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

func EthPerUSDT() (*big.Int, error) {
	priceFeedAddr := common.HexToAddress(rpc.EvmAddrs.EthPriceFeed)

	price, err := syncPriceLoop(priceFeedAddr)
	if err != nil {
		return nil, err
	}

	return price, nil
}
func BtcPerUSDT() (*big.Int, error) {
	priceFeedAddr := common.HexToAddress(rpc.EvmAddrs.BtcPriceFeed)

	price, err := syncPriceLoop(priceFeedAddr)
	if err != nil {
		return nil, err
	}

	return price, nil
}

func GasPriceOnChain() (*big.Int, error) {
	priceFeedAddr := common.HexToAddress(rpc.EvmAddrs.EthGasPriceFeed) // GASPRICE_PRICEFEED 有误,

	price, err := syncPriceLoop(priceFeedAddr)
	if err != nil {
		return nil, err
	}

	return price, nil
}

func syncPriceLoop(addr common.Address) (*big.Int, error) {
	// 小循环尽量保证接口不出错
	var err error
	for i := 0; i < 10; i++ {
		if i > 0 {
			logger.Debugf("%v failed to get gas price, try again", addr)
			time.Sleep(1 * time.Second)
		}
		price, errTemp := syncPrice(addr)
		if errTemp != nil {
			err = errTemp
			continue
		}
		return price, nil
	}
	return nil, err
}

func syncPrice(addr common.Address) (*big.Int, error) {
	ethRpc, err := rpc.NewEthRpc(true)
	if err != nil {
		return nil, err
	}
	defer ethRpc.Close()

	priceFeed, err := build.NewPriceFeed(addr, ethRpc.Client)
	if err != nil {
		return nil, err
	}

	price, err := priceFeed.PriceFeedCaller.LatestAnswer(&bind.CallOpts{})
	if err != nil {
		return nil, err
	}

	logger.Debug("syncPrice: price:", price, " addr:", addr)

	return price, nil
}
