package service

import (
	"collection-center/config"
	"collection-center/internal/btc"
	"collection-center/internal/logger"
	"collection-center/internal/rpc"
	"collection-center/library/constant"
	"collection-center/library/redis"
	"collection-center/library/request"
	"collection-center/library/utils"
	"errors"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	orgRedis "github.com/redis/go-redis/v9"
)

type LiquidityController struct {
	utils.Controller
}

func NewLiquidityController(ctx *gin.Context) *LiquidityController {
	c := &LiquidityController{}
	c.SetContext(ctx)
	return c
}

/*
	ETH: ETHBalance,

ETHLOCKED: liquidityLocked.eth.toString(),
USDT: USDTBalance,
USDTLOCKED: liquidityLocked.usdt.toString(),
BTC: BTCBalance,
BTCLOCKED: liquidityLocked.btc.toString()
*/
func (l *LiquidityController) FetchLiquidity() {
	ethCoreWalletAddress := config.Config().CoreWallet.EthWallet
	USDTContractAddress := config.Config().EvmAddress.UsdtErc20

	ethRpc, err := rpc.NewEthRpc()
	if err != nil {
		l.ResponseErr(err)
		return
	}
	// 使用 goroutine 并发获取数据
	var ETH, BTC, USDT, ETHLOCKED, USDTLOCKED, BTCLOCKED string
	wg := sync.WaitGroup{}
	wg.Add(4)

	go func() {
		defer wg.Done()
		ETH, err = ethRpc.BalanceOfETH(ethCoreWalletAddress)
		if err != nil {
			logger.Error("BalanceOfETH err:", err)
			return
		}
		ETH, err = utils.DecimalParse(ETH, 18)
		if err != nil {
			logger.Error("BalanceOfETH DecimalParse err:", err)
			return
		}
	}()

	go func() {
		defer wg.Done()
		USDT, err = ethRpc.BalanceOfERC20(ethCoreWalletAddress, USDTContractAddress)
		if err != nil {
			logger.Error("BalanceOfERC20 err:", err)
			return
		}
		USDT, err = utils.DecimalParse(USDT, 6)
		if err != nil {
			logger.Error("BalanceOfERC20 DecimalParse err:", err)
			return
		}
	}()

	go func() {
		defer wg.Done()
		BTC, err = btc.GetBalance(btc.BtcCoreWallet)
		if err != nil {
			logger.Error("btc GetBalance err:", err)
			return
		}
	}()

	go func() {
		defer wg.Done()
		ETHLOCKED, err = redis.GetChainData(constant.GetLqLockedKey(constant.CoinEth))
		if err != nil && !errors.Is(err, orgRedis.Nil) {
			logger.Error("redis GetChainData CoinEth err:", err)
			return
		}
		USDTLOCKED, err = redis.GetChainData(constant.GetLqLockedKey(constant.CoinUsdt))
		if err != nil && !errors.Is(err, orgRedis.Nil) {
			logger.Error("redis GetChainData CoinUsdt err:", err)
			return
		}
		BTCLOCKED, err = redis.GetChainData(constant.GetLqLockedKey(constant.CoinBtc))
		if err != nil && !errors.Is(err, orgRedis.Nil) {
			logger.Error("redis GetChainData CoinBtc err:", err)

			return
		}
	}()
	wg.Wait()
	if err != nil {
		l.ResponseErr(err)
		return
	}
	if ETHLOCKED == "" {
		ETHLOCKED = "0"
	}
	if USDTLOCKED == "" {
		USDTLOCKED = "0"
	}
	if BTCLOCKED == "" {
		BTCLOCKED = "0"
	}
	l.ResponseOk(map[string]string{
		"ETH":        ETH,
		"USDT":       USDT,
		"BTC":        BTC,
		"ETHLOCKED":  ETHLOCKED,
		"USDTLOCKED": USDTLOCKED,
		"BTCLOCKED":  BTCLOCKED,
	})
}

func (l *LiquidityController) FetchOut() {
	req := &request.OutReq{}
	err := l.Ctx.ShouldBind(req)
	if err != nil {
		l.ResponseErr(err)
		return
	}

	// 校验传入参数
	mode := strings.ToUpper(req.Mode)
	originToken := strings.ToUpper(req.Originaltoken)
	targetToken := strings.ToUpper(req.Targettoken)

	// 处理参数
	amount, _, err := CalculateOut(mode, originToken, req.Originaltokenamount, targetToken)
	if err != nil {
		l.ResponseErr(err)
		return
	}

	l.ResponseOk(map[string]string{
		"amount": amount,
	})
}
