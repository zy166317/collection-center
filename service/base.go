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

//func GetAuthPagePost(pageNumber int, pageSize int) *UserService.Page {
//	return &UserService.Page{
//		PageNumber: int64(pageNumber),
//		PageSize:   int64(pageSize),
//	}
//}
//func GetUserPageQuery(c *gin.Context) (*UserService.Page, error) {
//	pageNumber, _ := strconv.ParseInt(c.DefaultQuery("pageNumber", "1"), 10, 64)
//	pageSize, _ := strconv.ParseInt(c.DefaultQuery("pageSize", "10"), 10, 64)
//
//	return &UserService.Page{
//		PageNumber: pageNumber,
//		PageSize:   pageSize,
//	}, nil
//}
//func GetMessagePageQuery(c *gin.Context) *MessageService.Page {
//	pageNumber, _ := strconv.ParseInt(c.DefaultQuery("pageNumber", "1"), 10, 64)
//	pageSize, _ := strconv.ParseInt(c.DefaultQuery("pageSize", "10"), 10, 64)
//	return &MessageService.Page{
//		PageNumber: pageNumber,
//		PageSize:   pageSize,
//	}
//}
//func GetMessagePagePost(pageNumber int, pageSize int) *MessageService.Page {
//	return &MessageService.Page{
//		PageNumber: int64(pageNumber),
//		PageSize:   int64(pageSize),
//	}
//}
