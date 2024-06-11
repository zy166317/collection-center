package middleware

import (
	"github.com/afex/hystrix-go/hystrix"
	"github.com/gin-gonic/gin"
	"net/http"
)

const commandName = "GlobalHystrixConfig"

func InitHystrixConfig(config hystrix.CommandConfig) {
	hystrix.ConfigureCommand(commandName, config)
}

// HystrixMiddleware Gin 中间件
func HystrixMiddleware(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 调用 Hystrix 执行命令
		err := hystrix.Do(commandName, func() error {
			// 执行业务逻辑
			handler(c)
			return nil
		}, func(err error) error {
			// 后备逻辑
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "请求失败，请稍后再试"})
			return nil
		})

		// 如果 Hystrix 命令失败，则调用后备方法
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "请求失败，请稍后再试"})
		}
	}
}
