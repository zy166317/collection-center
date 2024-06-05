package service

import (
	cnt "collection-center/contract/constant"
	"collection-center/internal/logger"
	"collection-center/library/utils"
	"collection-center/service/price"
	"golang.org/x/xerrors"
	"math/big"
)

// UsdtToETHDec 可读USDT转精度ETH
func UsdtToETHDec(amount string) (*big.Int, error) {
	// decimals: 18
	decimals18, _ := utils.StrToBigFloat(cnt.DECIMALS_WEI)

	bFloatPrice, _, err := price.MultiTypeChainPrice()
	if err != nil {
		return nil, err
	}

	a, _ := utils.StrToBigFloat(amount)
	tta := new(big.Float).Mul(a, bFloatPrice.UsdtPerEth)

	valInt := new(big.Int)
	new(big.Float).Mul(tta, decimals18).Int(valInt)

	return valInt, nil
}

// EthDecToUSDTDec 精度ETH转精度USDT
func EthDecToUSDTDec(amount string) (*big.Int, error) {
	// decimals: 18
	decimals18, _ := utils.StrToBigFloat(cnt.DECIMALS_WEI)
	decimalsUSDT, _ := utils.StrToBigFloat(cnt.DECIMALS_USDT)

	bFloatPrice, _, err := price.MultiTypeChainPrice()
	if err != nil {
		return nil, err
	}

	val, err := utils.StrToBigFloat(amount)
	if err != nil {
		return nil, err
	}

	afterQuo := new(big.Float).Quo(val, decimals18)
	b := new(big.Float).Quo(afterQuo, bFloatPrice.UsdtPerEth)

	valInt := new(big.Int)
	new(big.Float).Mul(b, decimalsUSDT).Int(valInt)

	return valInt, nil
}

func ReadNumToDecNum(amount *big.Float, decLen int) (*big.Int, error) {
	var dec *big.Float
	switch decLen {
	case 6:
		dec, _ = utils.StrToBigFloat(cnt.DECIMALS_USDT)
	case 8:
		dec, _ = utils.StrToBigFloat(cnt.DECIMALS_EIGHT)
	case 18:
		dec, _ = utils.StrToBigFloat(cnt.DECIMALS_WEI)
	default:
		return nil, xerrors.New("Invalid decimals value")
	}

	// big.Float => big.Int
	covertNum := new(big.Int)
	new(big.Float).Mul(
		amount,
		dec,
	).Int(covertNum)

	return covertNum, nil
}

func DecNumToReadNum(amount *big.Int, decLen int) (*big.Float, error) {
	var dec *big.Float
	switch decLen {
	case 6:
		dec, _ = utils.StrToBigFloat(cnt.DECIMALS_USDT)
	case 8:
		dec, _ = utils.StrToBigFloat(cnt.DECIMALS_EIGHT)
	case 18:
		dec, _ = utils.StrToBigFloat(cnt.DECIMALS_WEI)
	default:
		return nil, xerrors.New("Invalid decimals value")
	}

	coverNum := new(big.Float).Quo(
		new(big.Float).SetInt(amount),
		dec,
	)

	return coverNum, nil
}

func CalculateOut(
	mode string,
	originToken string,
	otAmount string,
	targetToken string,
) (string, *big.Float, error) {
	var out string

	var amountOut *big.Float
	var fee *big.Float
	var pureAmount *big.Float
	var rate *big.Float
	var firstGasDef *big.Float

	ppg, err := price.GasPriceEth()
	if err != nil {
		return "", nil, err
	}

	bFloatPrice, _, err := price.MultiTypeChainPrice()
	if err != nil {
		logger.Error("MultiTypeChainPrice err:", err)
		return "", nil, err
	}

	inAmount, err := utils.StrToBigFloat(otAmount)
	if err != nil {
		return "", nil, err
	}

	if mode == "FIXED" {
		rate, err = utils.StrToBigFloat(cnt.FIXED_MODE_RATE)
	} else if mode == "FLOAT" {
		rate, err = utils.StrToBigFloat(cnt.FLOAT_MODE_RATE)
	} else {
		err = xerrors.New("Invalid mode:" + mode)
	}
	if err != nil {
		logger.Error("Invalid mode, err:", err)
		return "", nil, err
	}

	/*
		计算流程
		A Token => USDT => B Token
	*/
	if originToken == "ETH" && targetToken == "USDT" {
		pureAmount = new(big.Float).Mul(inAmount, bFloatPrice.EthPerUsdt)
		// 计算两笔成本
		gasLimit := new(big.Float).SetInt64(int64(cnt.GASLIMIT_ERC20 + cnt.GASLIMIT_ETH))
		fee = new(big.Float).Mul(gasLimit, ppg)

		firstGasDef, err = CalculateDefGas("ETH", ppg, bFloatPrice.UsdtPerEth)
		if err != nil {
			return "", nil, err
		}

	} else if originToken == "USDT" && targetToken == "ETH" {
		pureAmount = new(big.Float).Mul(inAmount, bFloatPrice.UsdtPerEth)
		// 计算4笔成本
		gasLimit := new(big.Float).SetInt64(int64(cnt.GASLIMIT_ERC20 + cnt.GASLIMIT_ETH*3))
		fee = new(big.Float).Mul(gasLimit, ppg)

		firstGasDef, err = CalculateDefGas("USDT", ppg, bFloatPrice.UsdtPerEth)
		if err != nil {
			return "", nil, err
		}

	} else if originToken == "BTC" && targetToken == "USDT" {
		pureAmount = new(big.Float).Mul(inAmount, bFloatPrice.BtcPerUsdt)
		// 第一笔成本(BTC) BTC => USDT
		bGAS, _ := utils.StrToBigFloat(cnt.DEFAULT_BTC_GAS)
		gas1 := new(big.Float).Mul(bGAS, bFloatPrice.BtcPerUsdt)
		// 第二笔成本(TOKEN)
		gasLimit := new(big.Float).SetInt64(int64(cnt.GASLIMIT_ERC20))
		gas2 := new(big.Float).Mul(gasLimit, ppg)
		// 转成USDT
		gas2 = new(big.Float).Mul(gas2, bFloatPrice.EthPerUsdt)

		fee = new(big.Float).Add(gas1, gas2)

		firstGasDef, err = CalculateDefGas("BTC", bGAS, bFloatPrice.UsdtPerBtc)
		if err != nil {
			return "", nil, err
		}

	} else if originToken == "USDT" && targetToken == "BTC" {
		pureAmount = new(big.Float).Mul(inAmount, bFloatPrice.UsdtPerBtc)
		// 第一笔成本(TOKEN)
		gasLimit := new(big.Float).SetInt64(int64(cnt.GASLIMIT_ERC20 + cnt.GASLIMIT_ETH*2))
		gas1 := new(big.Float).Mul(gasLimit, ppg)
		// 转成USDT
		gas1 = new(big.Float).Mul(gas1, bFloatPrice.EthPerUsdt)
		// 再转成BTC
		gas1 = new(big.Float).Mul(gas1, bFloatPrice.UsdtPerBtc)
		// 第二笔成本(BTC)
		gas2, _ := utils.StrToBigFloat(cnt.DEFAULT_BTC_GAS)
		fee = new(big.Float).Add(gas1, gas2)

		firstGasDef, err = CalculateDefGas("USDT", ppg, bFloatPrice.UsdtPerEth)
		if err != nil {
			return "", nil, err
		}

	} else if originToken == "ETH" && targetToken == "BTC" {
		pureAmount = new(big.Float).Mul(inAmount, bFloatPrice.EthPerUsdt)
		pureAmount = new(big.Float).Mul(pureAmount, bFloatPrice.UsdtPerBtc)

		// 第一笔成本(ETH) ETH => BTC
		gasLimit := new(big.Float).SetInt64(int64(cnt.GASLIMIT_ETH))
		gas1 := new(big.Float).Mul(gasLimit, ppg)
		// 转成USDT
		gas1 = new(big.Float).Mul(gas1, bFloatPrice.EthPerUsdt)
		// 再转成BTC
		gas1 = new(big.Float).Mul(gas1, bFloatPrice.UsdtPerBtc)
		// 第二笔成本(BTC)
		gas2, _ := utils.StrToBigFloat(cnt.DEFAULT_BTC_GAS)
		fee = new(big.Float).Add(gas1, gas2)

		firstGasDef, err = CalculateDefGas("ETH", ppg, bFloatPrice.UsdtPerEth)
		if err != nil {
			return "", nil, err
		}

	} else if originToken == "BTC" && targetToken == "ETH" {
		// BTC => USDT (pureAmount)
		pureAmount = new(big.Float).Mul(inAmount, bFloatPrice.BtcPerUsdt)
		pureAmount = new(big.Float).Mul(pureAmount, bFloatPrice.UsdtPerEth)

		// 第一笔成本(BTC) BTC => ETH
		bGAS, _ := utils.StrToBigFloat(cnt.DEFAULT_BTC_GAS)
		gas1 := new(big.Float).Mul(bGAS, bFloatPrice.BtcPerUsdt)
		gas1 = new(big.Float).Mul(gas1, bFloatPrice.UsdtPerEth)
		// 第二笔成本(ETH)
		gasLimit := new(big.Float).SetInt64(int64(cnt.GASLIMIT_ETH))
		gas2 := new(big.Float).Mul(gasLimit, ppg)

		fee = new(big.Float).Add(gas1, gas2)

		firstGasDef, err = CalculateDefGas("BTC", bGAS, bFloatPrice.UsdtPerBtc)
		if err != nil {
			return "", nil, err
		}

	} else {
		// 无效数据
		return "", nil, xerrors.New("Invalid in amount")
	}

	amountOut = new(big.Float).Mul(
		new(big.Float).Sub(pureAmount, fee),
		rate,
	)

	logger.Warnf("out amount:%.18f\n", amountOut)

	out = amountOut.String()

	return out, firstGasDef, nil
}

// 计算第一笔默认gas成本（以USDT计价）
func CalculateDefGas(originToken string, ppg *big.Float, floatPrice *big.Float) (*big.Float, error) {
	var gasCost *big.Float
	switch originToken {
	case "ETH":
		gas := new(big.Float).Mul(
			new(big.Float).SetInt64(int64(cnt.GASLIMIT_ETH)),
			ppg,
		)
		gasCost = new(big.Float).Quo(gas, floatPrice)

		break
	case "USDT":
		gas := new(big.Float).Mul(
			new(big.Float).SetInt64(int64(cnt.GASLIMIT_ERC20+cnt.GASLIMIT_ETH*2)),
			ppg,
		)
		// 转成USDT
		gasCost = new(big.Float).Quo(gas, floatPrice)

		break
	case "BTC":
		gas, _ := utils.StrToBigFloat(cnt.DEFAULT_BTC_GAS)
		gasCost = new(big.Float).Quo(gas, floatPrice)

		break
	default:
		return nil, xerrors.New("Invalid origin token")
	}

	return gasCost, nil
}
