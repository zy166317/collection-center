package middleware

import (
	"github.com/gin-gonic/gin"
)

// session 中间件
// 创建订单后, 生成 session, 有效期 7天, 用于用户查询订单状态, 一对多
// session key: defaultOrderSession
// session value: redis 中的 key, 有效期 7天
// 从 redis 中获取 orderIds, key 为 orderIdsStr --> session value

// SessionAuth 根据 session 验证 orderId
func SessionAuth() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}
