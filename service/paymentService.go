package service

import (
	"collection-center/internal/ecode"
	"collection-center/library/request"
)

func CreatePayment(req *request.CreatePaymentReq, merchantUid int64) error {
	//参数校验
	if req.CollectUid == 0 || req.ProjectUid == 0 || req.CreationChain == "" || req.CreationTokenSymbol == "" || req.ReturnUrl == "" || merchantUid == 0 {
		return ecode.IllegalParam
	}
	//校验项目是否支持chain和symbol

	//获取当前token对应U的价值
	return nil
}
