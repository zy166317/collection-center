package payVerify

import (
	"collection-center/service"
	"github.com/gin-gonic/gin"
)

func InitPayVerifyRoutes(engine *gin.Engine) {
	group := engine.Group("/verify")
	group.POST("/addhash", PayHash)
}

func PayHash(ctx *gin.Context) {
	service.NewPayVerifyController(ctx).AddPendingOrder()
}
