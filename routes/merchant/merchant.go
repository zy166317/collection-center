package merchant

import (
	"collection-center/service"
	"github.com/gin-gonic/gin"
)

// InitMerchantRoutes 初始化商家相关的路由配置。
func InitMerchantRoutes(engine *gin.Engine) {
	group := engine.Group("/merchant")
	group.POST("/sendVerifyCode", SendVerifyCode)
	group.POST("/register", RegisterMerchant)
	group.POST("/login", LoginMerchant)
}

// SendVerifyCode 发送商户注册邮箱验证代码。
func SendVerifyCode(ctx *gin.Context) {
	service.NewMerchantController(ctx).SendEmailVerifyCode()
}

// RegisterMerchant 处理商户注册请求。
func RegisterMerchant(ctx *gin.Context) {
	service.NewMerchantController(ctx).RegisterMerchant()
}

// LoginMerchant 处理商户登录请求。
func LoginMerchant(ctx *gin.Context) {
	service.NewMerchantController(ctx).LoginMerchant()
}
