package price

import (
	cnt "collection-center/contract/constant"
	"collection-center/internal/logger"
	"collection-center/library/redis"
	"collection-center/library/utils"
	"math/big"
)

type BigFloPrice struct {
	EthPerUsdt *big.Float
	BtcPerUsdt *big.Float
	UsdtPerEth *big.Float // 可读
	UsdtPerBtc *big.Float // 可读
}

type BigIntPrice struct {
	EthPerUsdt *big.Int
	BtcPerUsdt *big.Int
	UsdtPerEth *big.Int
	UsdtPerBtc *big.Int
}

// MultiTypeChainPrice 获取链上价格
func MultiTypeChainPrice() (BigFloPrice, BigIntPrice, error) {
	var BigFloPriceObj BigFloPrice
	var BigIntPriceObj BigIntPrice

	decimals8, _ := utils.StrToBigFloat(cnt.DECIMALS_EIGHT)
	decimals18, _ := utils.StrToBigFloat(cnt.DECIMALS_WEI)
	decimalsU, _ := utils.StrToBigFloat(cnt.DECIMALS_USDT)

	btcPrice, err := redis.GetChainData(cnt.BAPU)
	if err != nil {
		return BigFloPriceObj, BigIntPriceObj, err
	}
	bapu, err := utils.StrToBigFloat(btcPrice)
	if err != nil {
		return BigFloPriceObj, BigIntPriceObj, err
	}
	bapu = new(big.Float).Quo(bapu, decimals8)
	BigFloPriceObj.BtcPerUsdt = bapu

	bBapu18 := new(big.Int)
	new(big.Float).Mul(bapu, decimals18).Int(bBapu18)
	BigIntPriceObj.BtcPerUsdt = bBapu18

	ethPrice, err := redis.GetChainData(cnt.EAPU)
	if err != nil {
		return BigFloPriceObj, BigIntPriceObj, err
	}
	eapu, err := utils.StrToBigFloat(ethPrice)
	if err != nil {
		return BigFloPriceObj, BigIntPriceObj, err
	}
	eapu = new(big.Float).Quo(eapu, decimals8)
	BigFloPriceObj.EthPerUsdt = eapu

	bEapu18 := new(big.Int)
	new(big.Float).Mul(eapu, decimals18).Int(bEapu18)
	BigIntPriceObj.EthPerUsdt = bEapu18

	floatOne := big.NewFloat(1)
	usdtPerBTC := new(big.Float).Quo(floatOne, bapu)
	BigFloPriceObj.UsdtPerBtc = usdtPerBTC

	bUsdtPerBTCU := new(big.Int)
	new(big.Float).Mul(usdtPerBTC, decimalsU).Int(bUsdtPerBTCU)
	BigIntPriceObj.UsdtPerBtc = bUsdtPerBTCU

	usdtPerETH := new(big.Float).Quo(floatOne, eapu)
	BigFloPriceObj.UsdtPerEth = usdtPerETH

	bUsdtPerETHU := new(big.Int)
	new(big.Float).Mul(usdtPerETH, decimalsU).Int(bUsdtPerETHU)
	BigIntPriceObj.UsdtPerEth = bUsdtPerETHU

	return BigFloPriceObj, BigIntPriceObj, err
}

func GasPriceEth() (*big.Float, error) {
	//// decimals: 18
	decimals18, _ := utils.StrToBigFloat(cnt.DECIMALS_WEI)

	gasPrice, err := redis.GetChainData(cnt.PPG)
	if err != nil {
		return nil, err
	}
	logger.Debug("GasPriceEth: gasPrice:", gasPrice)
	// gas price(精度18)
	ppg, err := utils.StrToBigFloat(gasPrice)
	if err != nil {
		return nil, err
	}

	// 单位ETH可读
	ppg = new(big.Float).Quo(ppg, decimals18)
	if err != nil {
		return nil, err
	}

	return ppg, nil
}
