package middleware

import (
	"collection-center/internal/logger"
	"collection-center/library/response"
	"github.com/gin-gonic/gin"
	"net/http"
	"runtime"
	"runtime/debug"
)

func Recover(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			//打印错误堆栈信息
			logger.Error("panic: %v\n", r)
			switch r.(type) {
			case runtime.Error, error:
				// 运行时错误
				err := r.(error)
				debug.PrintStack()
				language := c.GetHeader("Language")
				c.AbortWithStatusJSON(http.StatusOK, response.ResErr(err, language))
			default:
				// 非运行时错误
			}
			c.Next()
		}
	}()
	//加载完 defer recover，继续后续接口调用
	c.Next()
}
