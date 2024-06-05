package service

import (
	"collection-center/library/utils"
	"github.com/gin-gonic/gin"
)

type PayVerifyController struct {
	utils.Controller
}

func NewPayVerifyController(ctx *gin.Context) *PayVerifyController {
	c := &PayVerifyController{}
	c.SetContext(ctx)
	return c
}

func (p *PayVerifyController) AddHash() {

}
