package payVerify

import (
	"collection-center/middleware"
	"collection-center/service"
	"github.com/gin-gonic/gin"
)

func InitPayVerifyRoutes(engine *gin.Engine) {
	group := engine.Group("/verify")
	group.POST("/addhash", middleware.SessionAuth(), PayHash)
}

func PayHash(ctx *gin.Context) {
	service.NewPayVerifyController(ctx).AddHash()
}
