package priceFeed

import (
	"collection-center/service"
	"github.com/gin-gonic/gin"
)

func InitPriceFeedRoutes(engine *gin.Engine) {
	group := engine.Group("/price")
	group.GET("/eth", EthPricefeed)
	group.GET("/btc", BtcPricefeed)
	group.GET("/gas", GasPricefeed)
	// 为了满足之前版本的接口，保留以下接口
	engine.GET("/ethusdt", EthUsdtPricefeed)
	engine.GET("/btcusdt", BtcUsdtPricefeed)

	//engine.GET("/ethusdt", EthPricefeed)
	//engine.GET("/btcusdt", BtcPricefeed)
}

// EthPricefeed
// @Summary 获取 eth pricefeed
// @Description 获取 eth pricefeed
// @Tags [PriceFeed]ETH
// @Param Language header string false "用户语言 CN 或EN 不填默 CN"
// @Accept application/json
// @Produce application/json
// @Success 200 {string} json "{'code':0,'message':'Success','data':{'price':'154492827441'}}"
// @Router /price/eth [GET]
func EthPricefeed(ctx *gin.Context) {
	service.NewPricefeedController(ctx).FetchEthPricefeed()
}

// BtcPricefeed
// @Summary 获取 btc pricefeed
// @Description 获取 btc pricefeed
// @Tags [PriceFeed]btc
// @Param Language header string false "用户语言 CN 或EN 不填默 CN"
// @Accept application/json
// @Produce application/json
// @Success 200 {string} json "{'code':0,'message':'Success','data':{'price':'2821462000000'}}"
// @Router /price/btc [GET]
func BtcPricefeed(ctx *gin.Context) {
	service.NewPricefeedController(ctx).FetchBtcPricefeed()
}

// GasPricefeed
// @Summary 获取 eth gas pricefeed
// @Description 获取 eth gas pricefeed
// @Tags [PriceFeed]eth gas
// @Param Language header string false "用户语言 CN 或EN 不填默 CN"
// @Accept application/json
// @Produce application/json
// @Success 200 {string} json "{'code':0,'message':'Success','data':{'price':'154492827441'}}"
// @Router /price/gas [GET]
func GasPricefeed(ctx *gin.Context) {
	service.NewPricefeedController(ctx).FetchGasPricefeed()
}

// EthUsdtPricefeed
// @Summary 获取 eth pricefeed
// @Description 获取 eth pricefeed
// @Tags [PriceFeed]ETH
// @Param Language header string false "用户语言 CN 或EN 不填默 CN"
// @Accept application/json
// @Produce application/json
// @Success 200 {string} json "{'code':0,'message':'Success','data':{'price':'154492827441'}}"
// @Router /ethusdt [GET]
func EthUsdtPricefeed(ctx *gin.Context) {
	service.NewPricefeedController(ctx).FetchEthUsdtPricefeed()
}

// BtcUsdtPricefeed
// @Summary 获取 btc pricefeed
// @Description 获取 btc pricefeed
// @Tags [PriceFeed]btc
// @Param Language header string false "用户语言 CN 或EN 不填默 CN"
// @Accept application/json
// @Produce application/json
// @Success 200 {string} json "{'code':0,'message':'Success','data':{'price':'2821462000000'}}"
// @Router /btcusdt [GET]
func BtcUsdtPricefeed(ctx *gin.Context) {
	service.NewPricefeedController(ctx).FetchBtcUsdtPricefeed()
}
