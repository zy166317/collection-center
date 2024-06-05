package service

import (
	"collection-center/contract/constant"
	"collection-center/library/redis"
	"collection-center/library/utils"
	"collection-center/service/price"
	"github.com/gin-gonic/gin"
	orgRedis "github.com/redis/go-redis/v9"
	"math/big"
)

type PricefeedController struct {
	utils.Controller
}

func NewPricefeedController(ctx *gin.Context) *PricefeedController {
	c := &PricefeedController{}
	c.SetContext(ctx)
	return c
}

func (b *PricefeedController) FetchEthPricefeed() {
	ethPrice, err := price.EthPerUSDT()
	if err != nil {
		b.ResponseErr(err)
		return
	}

	// 待升级
	// decimals: 8
	decimals8, _ := utils.StrToBigFloat(constant.DECIMALS_EIGHT)
	ePrice := new(big.Float).SetInt(ethPrice)
	ePrice = new(big.Float).Quo(ePrice, decimals8)

	b.ResponseOk(map[string]string{
		"price": ePrice.String(),
	})
}

func (b *PricefeedController) FetchBtcPricefeed() {
	btcPrice, err := price.BtcPerUSDT()
	if err != nil {
		b.ResponseErr(err)
		return
	}

	// 待升级
	// decimals: 8
	decimals8, _ := utils.StrToBigFloat(constant.DECIMALS_EIGHT)
	bPrice := new(big.Float).SetInt(btcPrice)
	bPrice = new(big.Float).Quo(bPrice, decimals8)

	b.ResponseOk(map[string]string{
		"price": bPrice.String(),
	})
}

func (b *PricefeedController) FetchGasPricefeed() {
	ethPrice, err := price.GasPriceOnChain()
	if err != nil {
		b.ResponseErr(err)
		return
	}

	b.ResponseOk(map[string]string{
		"price": ethPrice.String(),
	})
}

func (b *PricefeedController) FetchEthUsdtPricefeed() {
	// 从 redis 中获取 eth 单价
	priceStr, err := redis.GetChainData(constant.EAPU)
	if err != nil && err != orgRedis.Nil {
		b.ResponseErr(err)
		return
	}
	if err == orgRedis.Nil {
		// 从链上获取 eth 单价
		ethPerUSDT, err := price.EthPerUSDT()
		if err != nil {
			b.ResponseErr(err)
			return
		}
		// 存入 redis
		_, err = redis.InsertChainData(constant.EAPU, ethPerUSDT.String())
		if err != nil {
			b.ResponseErr(err)
			return
		}
		priceStr = ethPerUSDT.String()
	}
	epu, err := utils.StrToBigFloat(priceStr)
	if err != nil {
		b.ResponseErr(err)
		return
	}
	// 转换成 一个usdt对应的eth数量
	fPrice := new(big.Float).Quo(big.NewFloat(100000000), epu)
	resp, err := utils.AsStringFromFloat(18, fPrice)
	if err != nil {
		b.ResponseErr(err)
		return
	}
	b.ResponseOk(map[string]string{
		"price": resp,
	})
}

func (b *PricefeedController) FetchBtcUsdtPricefeed() {
	// 从 redis 中获取 btc 单价
	priceStr, err := redis.GetChainData(constant.BAPU)
	if err != nil && err != orgRedis.Nil {
		b.ResponseErr(err)
		return
	}
	if err == orgRedis.Nil {
		// 从链上获取 btc 单价
		btcPerUSDT, err := price.BtcPerUSDT()
		if err != nil {
			b.ResponseErr(err)
			return
		}
		// 存入 redis
		_, err = redis.InsertChainData(constant.BAPU, btcPerUSDT.String())
		if err != nil {
			b.ResponseErr(err)
			return
		}
		priceStr = btcPerUSDT.String()
	}
	bapu, err := utils.StrToBigFloat(priceStr)
	if err != nil {
		b.ResponseErr(err)
		return
	}
	// 转换成 一个usdt对应的btc数量
	fPrice := new(big.Float).Quo(big.NewFloat(100000000), bapu)
	resp, err := utils.AsStringFromFloat(18, fPrice)
	if err != nil {
		b.ResponseErr(err)
		return
	}
	b.ResponseOk(gin.H{
		"price": resp,
	})
}
