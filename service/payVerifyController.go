package service

import (
	"collection-center/library/request"
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

func (p *PayVerifyController) AddPendingOrder() {
	req := &request.PendingOrderReq{}
	err := p.Ctx.ShouldBind(req)
	if err != nil {
		p.ResponseErr(err)
		return
	}
	ip := p.Ctx.RemoteIP()
	req.NotifyIp = ip
	err = AddPendingTonOrder(req)
	if err != nil {
		p.ResponseErr(err)
		return
	}
	p.ResponseOk(nil)
}
