package service

import (
	cnt "collection-center/contract/constant"
	"collection-center/internal/btc"
	"collection-center/internal/signClient"
	"collection-center/library/redis"
	"collection-center/library/utils"
	"collection-center/service/price"
	"fmt"
	"math/big"
	"testing"
)

func init() {
	redis.SetRedis(&redis.RedisConfig{
		Addr:         "192.168.8.63:6379",
		Auth:         "orca_redis",
		DB:           1,
		DialTimeout:  0,
		ReadTimeout:  0,
		WriteTimeout: 0,
	})
	//btc.BtcRpcList = append(btc.BtcRpcList, "btc.getblock.io/c5049a57-bfc3-4be2-a66e-2844d0464825/testnet/")
	signClient.SignerConfig.Host = "192.168.8.63"
	signClient.SignerConfig.Port = "8080"
	signClient.SignerConfig.TlsPemPath = "../../resources/grpc-server-cert.pem"
	signClient.SignerConfig.User = "orca_off_signer"
	signClient.SignerConfig.Pass = "orca_6b9c1bb0f29f80a4fa759d8af2d26dd2"
	btc.InitBtcd(true)
}

func TestFirstGasCost(t *testing.T) {
	ppg, _ := price.GasPriceEth()
	bFloatPrice, _, _ := price.MultiTypeChainPrice()

	gas1, _ := CalculateDefGas("ETH", ppg, bFloatPrice.UsdtPerEth)
	fmt.Printf("Gas1:%.18f\n", gas1)
	gas2, _ := CalculateDefGas("USDT", ppg, bFloatPrice.UsdtPerEth)
	fmt.Printf("Gas2:%.18f\n", gas2)
	bGAS, _ := utils.StrToBigFloat(cnt.DEFAULT_BTC_GAS)
	gas3, _ := CalculateDefGas("BTC", bGAS, bFloatPrice.UsdtPerBtc)
	fmt.Printf("Gas3:%.18f\n", gas3)

}

func TestCalculateOut(t *testing.T) {
	out, gas, _ := CalculateOut("FIXED", "USDT", "500", "ETH")
	fmt.Printf("out:%v\n", out)
	fmt.Printf("gas cost:%v USDT\n", gas)

	//CalculateOut("FIXED", "ETH", "20", "BTC")
	//CalculateOut("FIXED", "ETH", "1", "USDT")
	//CalculateOut("FIXED", "USDT", "200", "BTC")

	//bFloatPrice, _, err := price.MultiTypeChainPrice()
	//if err != nil {
	//	logger.Error("MultiTypeChainPrice err:", err)
	//	t.Error(err)
	//}
	//inAmount, err := utils.StrToBigFloat("200")
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
	//rate, err := utils.StrToBigFloat(cnt.FIXED_MODE_RATE)
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
	//ppg, err := price.GasPriceEth()
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
	//
	//pureAmount := new(big.Float).Mul(inAmount, bFloatPrice.UsdtPerBtc)
	//fmt.Printf("pureAmount:%.18f\n", pureAmount)
	//// 第一笔成本(TOKEN)
	//gasLimit := new(big.Float).SetInt64(int64(cnt.GASLIMIT_ERC20))
	//gas1 := new(big.Float).Mul(gasLimit, ppg)
	//fmt.Printf("ppg-1:%.18f\n", ppg)
	//fmt.Printf("gas1-1:%.18f\n", gas1)
	//// 转成USDT
	//gas1 = new(big.Float).Mul(gas1, bFloatPrice.EthPerUsdt)
	//fmt.Printf("gas1-2:%.18f\n", gas1)
	//// 再转成BTC
	//gas1 = new(big.Float).Mul(gas1, bFloatPrice.UsdtPerBtc)
	//fmt.Printf("gas1-3:%.18f\n", gas1)
	//// 第二笔成本(BTC)
	//btcGas, err := btc.Estimatefee("mqdEyyExvwmDyzXRaxSo6Yg8c5yaSKHTL1", "tb1quq809m29ggx2t2e0nv0vewzztz78z6u304pfmc", pureAmount.Text('f', 8))
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
	//gas2, _ := utils.StrToBigFloat(btcGas)
	//fmt.Printf("btcGas:%s\n", btcGas)
	//fee := new(big.Float).Add(gas1, gas2)
	//amountOut := new(big.Float).Mul(
	//	new(big.Float).Sub(pureAmount, fee),
	//	rate,
	//)
	//fmt.Printf("fee:%.18f\n", fee)
	//fmt.Printf("amountOut:%.18f\n", amountOut)
	////CalculateOut("FIXED", "BTC", "30", "USDT")
}

func TestUsdtToETHDec(t *testing.T) {
	ret, err := UsdtToETHDec("1550")
	if err != nil {
		t.Error(err)
	}

	a, _ := utils.StrToBigFloat(ret.String())
	b, _ := utils.StrToBigFloat(cnt.DECIMALS_WEI)
	ethVal := new(big.Float).Quo(a, b)

	fmt.Printf("USDT To ETH:%d\n", ret)
	fmt.Printf("Format to ETH:%.18f\n", ethVal)
}

func TestEthDecToUSDTDec(t *testing.T) {
	ret, err := EthDecToUSDTDec(cnt.DECIMALS_WEI)
	if err != nil {
		t.Error(err)
	}

	a, _ := utils.StrToBigFloat(ret.String())
	b, _ := utils.StrToBigFloat(cnt.DECIMALS_USDT)
	val := new(big.Float).Quo(a, b)

	fmt.Printf("ETH To USDT:%d\n", ret)
	fmt.Printf("Format to USDT:%.18f\n", val)
}
