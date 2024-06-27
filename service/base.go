package service

import (
	"collection-center/library/request"
	"collection-center/library/utils"
	"collection-center/service/db/dao"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	utils.Controller
}

func NewController(ctx *gin.Context) *Controller {
	c := &Controller{}
	c.SetContext(ctx)
	return c
}

func (c *Controller) Ping() (err error) {
	//ping := &TestService.PingContent{ // grpc 测试
	//	Msg: "123",
	//}
	//c.Response(err, gin.H{"msg": ping.Msg})
	c.ResponseOk("OK")
	return
}
func GetPage(pageNumber int, pageSize int) dao.Page {
	return dao.Page{
		PageNumber: pageNumber,
		PageSize:   pageSize,
	}
}
func ParsePage(page request.Page) dao.Page {
	return dao.Page{
		PageNumber: page.PageNumber,
		PageSize:   page.PageSize,
	}
}
