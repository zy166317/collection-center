package order

import (
	"collection-center/middleware"
	"collection-center/service"
	"github.com/gin-gonic/gin"
)

func InitOrderRoutes(engine *gin.Engine) {
	group := engine.Group("/order")
	group.GET("/test", TestOrder)
	group.GET("/details", middleware.SessionAuth(), OrderDetail)
	group.GET("/historyorder", History)
	// Brief 返回订单 Summary 简要信息
	group.GET("/brief", Brief)

	group.POST("/generateorder", middleware.RateLimit(10, 60, "/order/generateorder"), Generate)
	group.POST("/refreshorder", middleware.SessionAuth(), Refresh)
	group.POST("/refund", middleware.SessionAuth(), Refund)
	//新增支付校验
}

// TestOrder
// @Summary 测试
// @Description 测试
// @Tags [Order]TestOrder
// @Param Language header string false "用户语言 CN 或EN 不填默 CN"
// @Accept application/json
// @Produce application/json
// @Success 200 {string} json "{'code':0,'message':'Success','data':{'status': true}}"
// @Router /order/test [GET]
func TestOrder(ctx *gin.Context) {
	service.NewOrderController(ctx).TestOrder()
}

// OrderDetail
// @Summary 获取 order 详情
// @Description 根据 order_id 获取 order 详情
// @Tags [Order]Order
// @Param Language header string false "用户语言 CN 或EN 不填默 CN"
// @Accept application/json
// @Produce application/json
// @Param id query string true "订单ID"
// @Success 200 {string} json "{'code':0,'message':'Success','data':{'todo':'todo'}}"
// @Router /order/detail [GET]
func OrderDetail(ctx *gin.Context) {
	service.NewOrderController(ctx).OrderDetails()
}

// History
// @Summary 获取 order 历史记录
// @Description 获取 order 历史记录 - 最新10条
// @Tags [Order]Order
// @Param Language header string false "用户语言 CN 或EN 不填默 CN"
// @Accept application/json
// @Produce application/json
// @Success 200 {string} json "{'code':0,'message':'Success','data':[service.HistoryOrder]}"
// @Router /order/historyorder [GET]
func History(ctx *gin.Context) {
	service.NewOrderController(ctx).HistoryOrder()
}

// Generate
// @Summary 创建订单
// @Description 创建订单
// @Tags [Order]Generate
// @Param Language header string false "用户语言 CN 或EN 不填默 CN"
// @Accept application/json
// @Param object body request.OrderReq true "{'mode':”, 'originaltoken':”, 'originaltokenamount':”,'targettoken':”,'targettokenamount':”,'userreceiveaddress':”,'email':”}"
// @Produce application/json
// @Success 200 {string} json "{'code':0,'message':'Success','data':{'Order':{}}}"
// @Router /order/generateorder [POST]
func Generate(ctx *gin.Context) {
	service.NewOrderController(ctx).GenOrder()
}

// Refund
// @Summary 订单退款
// @Description 订单退款
// @Tags [Order]Refund
// @Param Language header string false "用户语言 CN 或EN 不填默 CN"
// @Accept application/json
// @Param object body request.RefundReq true "{'id':”, 'refundaddress':”, 'email':”}"
// @Produce application/json
// @Success 200 {string} json "{'code':0,'message':'Success','data':dao.Refund}"
// @Router /order/refund [POST]
func Refund(ctx *gin.Context) {
	service.NewOrderController(ctx).RefundOrder()
}

// TODO api impl and swag
func Refresh(ctx *gin.Context) {
	service.NewOrderController(ctx).RefreshOrder()
}

// Brief
// @Summary 获取 order sum 简要信息
// @Description 获取 order sum 简要信息
// @Tags [Order]Brief
// @Param Language header string false "用户语言 CN 或EN 不填默 CN"
// @Accept application/json
// @Produce application/json
// @Success 200 {string} json "{'code':0,'message':'Success','data':{['allTime':200.3,'past24':20.3,'past7':10.3,'past30':30.3]}}"
// @Router /order/brief [GET]
func Brief(ctx *gin.Context) {
	service.NewOrderController(ctx).BriefOrder()
}
