package routes

import (
	"collection-center/service"
	"fmt"
	"github.com/gin-gonic/gin"
)

func InitRoutes(engine *gin.Engine) {
	engine.GET("/test/ping", ping)
	//engine.GET("/test/mauth/ping", middleware.CheckAuth("Test1", "Test"), middleware.CheckPermission("TEST"), authPing)
	//engine.POST("/test/getFile", getFile)
	//engine.POST("/test/oss", getOss)
}

// @Summary oss
// @Description oss
// @Tags [测试]oss
// @Param Language header string false "用户语言 CN 或EN 不填默 CN"
// @Accept application/json
// @Produce application/json
// @Param query query string true "aa"
// @Success 200 {string} json "{"RetCode":0,"UserInfo":{},"Action":"GetOneUserResponse"}"
// @Router /test/oss [POST]
//func getOss(ctx *gin.Context) {
//	err, s, s2, s3 := ossutil.GetTempInfo()
//	if err != nil {
//		fmt.Printf(err.Error())
//	}
//	fmt.Println(s, s2, s3)
//}

// @Summary 下载zip文件
// @Description 下载zip文件
// @Tags [测试]下载zip文件
// @Param Language header string false "用户语言 CN 或EN 不填默 CN"
// @Accept application/json
// @Produce application/json
// @Param query query string true "测试参数"
// @Success 200 {string} json "{"RetCode":0,"UserInfo":{},"Action":"GetOneUserResponse"}"
// @Router /test/ping [POST]
//func getFile1(ctx *gin.Context) {
//	//utils.Demo1(ctx.Writer)
//}

// @Summary 上传文件/下载文件 测试
// @Description
// @Tags file
// @Accept multipart/form-data
// @Param file formData file true "file"
// @Produce  json
// @Success 200
// @Router /test/getFile [post]
//func getFile(ctx *gin.Context) {
//
//}

// @Summary 测试接口Summary
// @Description 测试接口Description
// @Tags [测试]测试接口
// @Param Language header string false "用户语言 CN 或EN 不填默 CN"
// @Accept application/json
// @Produce application/json
// @Param query query string true "测试参数"
// @Success 200 {string} json "{"OK"}"
// @Router /test/ping [get]
func ping(ctx *gin.Context) {
	query, _ := ctx.GetQuery("query")
	fmt.Println("query:" + query)
	service.NewController(ctx).Ping()
}

// @Summary 测试授权
// @Description 测试接口Description
// @Tags [测试]测试接口
// @Param Token header string false "用户令牌"
// @Param Language header string false "用户语言 CN 或EN 不填默 CN"
// @Accept application/json
// @Produce application/json
// @Param query query string true "测试参数"
// @Success 200 {string} json "{"RetCode":0,"UserInfo":{},"Action":"GetOneUserResponse"}"
// @Router /test/mauth/ping [get]
//func authPing(ctx *gin.Context) {
//	query, _ := ctx.GetQuery("query")
//	fmt.Println("query:" + query)
//	service.NewController(ctx).Ping()
//}
