package project

import (
	"collection-center/middleware"
	"collection-center/service"
	"github.com/gin-gonic/gin"
)

// InitProjectRoutes 初始化项目相关的路由配置。
func InitProjectRoutes(engine *gin.Engine) {
	group := engine.Group("/project", middleware.CheckAuth())
	group.POST("/create", middleware.CheckAuth(), CreateProject)
	group.POST("/addTokenInfo", middleware.CheckAuth(), AddTokenInfo)
	group.POST("/updateProjectInfo", middleware.CheckAuth(), UpdateProject)
	group.POST("/updateCollectRate", middleware.CheckAuth(), UpdateCollectRate)
	group.POST("/updateCollectAddress", middleware.CheckAuth(), UpdateCollectAddress)
	group.POST("/freezeProject", middleware.CheckAuth(), FreezeProject)
}

// CreateProject 创建项目
func CreateProject(ctx *gin.Context) {
	service.NewProjectController(ctx).CreateProject()
}

// AddTokenInfo 添加token信息
func AddTokenInfo(ctx *gin.Context) {
	service.NewProjectController(ctx).AddTokenInfo()
}

// UpdateProject 修改项目信息
func UpdateProject(ctx *gin.Context) {
	service.NewProjectController(ctx).UpdateProjectInfo()
}

// UpdateCollectRate 修改收款汇率
func UpdateCollectRate(ctx *gin.Context) {
	service.NewProjectController(ctx).UpdateCollectRate()
}

// UpdateCollectAddress 修改收款钱包地址
func UpdateCollectAddress(ctx *gin.Context) {
	service.NewProjectController(ctx).UpdateCollectAddress()
}

// FreezeProject 冻结项目
func FreezeProject(ctx *gin.Context) {
	service.NewProjectController(ctx).FreezeProjectReq()
}
