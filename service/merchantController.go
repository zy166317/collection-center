package service

import (
	"collection-center/internal/ecode"
	"collection-center/library/request"
	"collection-center/library/utils"
	"github.com/gin-gonic/gin"
)

type MerchantController struct {
	utils.Controller
}

func NewMerchantController(ctx *gin.Context) *MerchantController {
	c := &MerchantController{}
	c.SetContext(ctx)
	return c
}

func (u *MerchantController) SendEmailVerifyCode() {
	//判断邮箱是否已经注册
	req := &request.SendVerifyCodeReq{}
	err := u.Ctx.ShouldBind(req)
	if err != nil {
		u.ResponseWrapErr(ecode.IllegalParam, err)
		return
	}
	err = SendVerifyCode(req)
	if err != nil {
		u.ResponseErr(err)
		return
	}
	u.ResponseOk(map[string]interface{}{})
}

func (u *MerchantController) RegisterMerchant() {
	req := &request.CreateMerchantReq{}
	err := u.Ctx.ShouldBind(req)
	if err != nil {
		u.ResponseWrapErr(ecode.IllegalParam, err)
		return
	}
	//注册逻辑
	err = RegisterMerchant(req)
	if err != nil {
		u.ResponseErr(err)
		return
	}
	u.ResponseOk(map[string]interface{}{})
}

func (u *MerchantController) LoginMerchant() {
	req := &request.LoginMerchantReq{}
	err := u.Ctx.ShouldBind(req)
	if err != nil {
		u.ResponseWrapErr(ecode.IllegalParam, err)
		return
	}
	//注册逻辑
	token, err := LoginMerchant(req)
	if err != nil {
		u.ResponseErr(err)
		return
	}
	u.ResponseOk(map[string]interface{}{
		"token": token,
	})
}
