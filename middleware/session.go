package middleware

import (
	"collection-center/config"
	"collection-center/internal/ecode"
	"collection-center/library/response"
	"collection-center/service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// session 中间件
// 创建订单后, 生成 session, 有效期 7天, 用于用户查询订单状态, 一对多
// session key: defaultOrderSession
// session value: redis 中的 key, 有效期 7天
// 从 redis 中获取 orderIds, key 为 orderIdsStr --> session value

// SessionAuth 根据 session 验证 orderId
func SessionAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果是本地调用, 不需要验证
		clientIP := c.ClientIP()
		if clientIP == "::1" {
			clientIP = "127.0.0.1"
		}
		if clientIP == "127.0.0.1" {
			c.Next()
			return
		}
		// debug 下不需要验证
		if config.Config().Api.Debug {
			c.Next()
			return
		}
		// 获取 session
		session := sessions.Default(c)
		language := c.GetHeader("Language")
		_, orderIds, err := service.GetSessionValues(session)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, response.ResErr(err, language))
			return
		}

		if orderIds == nil {
			c.AbortWithStatusJSON(http.StatusOK, response.ResErr(ecode.AccessDenied, language))
			return
		}

		queryID := c.Query("id")
		id, err := strconv.ParseInt(queryID, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, ecode.IllegalParam)
			return
		}

		// 验证 orderId
		var check bool
		for _, orderID := range orderIds {
			if orderID == id {
				check = true
				break
			}
		}

		if !check {
			c.AbortWithStatusJSON(http.StatusOK, ecode.AccessDenied)
			return
		}
		c.Next()
	}
}
