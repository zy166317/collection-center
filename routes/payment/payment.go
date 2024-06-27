package payment

import "github.com/gin-gonic/gin"

// InitPaymentRoutes 初始化项目相关的路由配置。
func InitPaymentRoutes(engine *gin.Engine) {
	group := engine.Group("/payment")
	group.POST("/createProject", CreatePayment)
}

func CreatePayment(ctx *gin.Context) {

}
