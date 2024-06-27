package service

import (
	"collection-center/internal/ecode"
	"collection-center/library/request"
	"collection-center/library/utils"
	"github.com/gin-gonic/gin"
)

type PaymentController struct {
	utils.Controller
}

func NewPaymentController(ctx *gin.Context) *PaymentController {
	c := &PaymentController{}
	c.SetContext(ctx)
	return c
}

func (p *PaymentController) CreatePayment() {
	req := &request.CreatePaymentReq{}
	err := p.Ctx.ShouldBind(req)
	if err != nil {
		p.ResponseWrapErr(ecode.IllegalParam, err)
		return
	}
	value, _ := p.Ctx.Get("uid")
	err = CreatePayment(req, value.(int64))
}
