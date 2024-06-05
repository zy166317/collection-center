package liquidity

import (
	"collection-center/service"
	"github.com/gin-gonic/gin"
)

func InitLiquidity(engine *gin.Engine) {
	group := engine.Group("/liquidity")
	group.GET("/liquidity", Liquidity)
	group.GET("/out", Out)
}

// Liquidity
// @Summary 获取 Liquidity
// @Description 获取 Liquidity
// @Tags [Liquidity]Liquidity
// @Param Language header string false "用户语言 CN 或EN 不填默 CN"
// @Accept application/json
// @Produce application/json
// @Success 200 {string} json "{'code':0,'message':'Success','data':{'vol':'661231211'}}"
// @Router /liquidity/liquidity [GET]
func Liquidity(ctx *gin.Context) {
	service.NewLiquidityController(ctx).FetchLiquidity()
}

// Out
// @Summary 获取 Liquidity Out
// @Description 获取 Liquidity Out
// @Tags [Liquidity]Liquidity Out
// @Param Language header string false "用户语言 CN 或EN 不填默 CN"
// @Accept application/json
// @Produce application/json
// @Param object body request.OutReq true "{'mode':”, 'originaltoken':”, 'originaltokenamount':”,'targettoken':”}"
// @Success 200 {string} json "{'code':0,'message':'Success','data':{'amount':'661231211'}}"
// @Router /liquidity/out [GET]
func Out(ctx *gin.Context) {
	service.NewLiquidityController(ctx).FetchOut()
}
