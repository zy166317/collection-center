package server

import (
	"collection-center/internal/logger"
	"collection-center/library/constant"
	"collection-center/library/redis"
	"collection-center/middleware"
	"collection-center/routes"
	"collection-center/routes/merchant"
	"collection-center/routes/project"
	"errors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"net/http"
	"strconv"
)

var (
	server *http.Server
)

func HttpServer() *http.Server {
	return server
}
func NewHttpServer(port int, isDebug bool) *http.Server {
	engine := gin.Default()
	if isDebug {
		engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
	engine.Use(middleware.Recover)
	engine.Use(middleware.CrossDomain())
	engine.Use(sessions.Sessions(constant.DefaultOrderSession, redis.GetSessionStore()))
	//engine.Use(middleware.TraceLog(nil, nil))
	//engine.Use(middleware.CheckAuth())
	routes.InitRoutes(engine)
	// 路由逻辑
	//payVerify.InitPayVerifyRoutes(engine)
	merchant.InitMerchantRoutes(engine)
	project.InitProjectRoutes(engine)
	addr := ":" + strconv.Itoa(port)
	srv := &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("listen: %s\n", err)
		}
	}()
	server = srv
	return server
}
