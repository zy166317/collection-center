package middleware

import (
	"collection-center/config"
	"collection-center/internal/ecode"
	"collection-center/internal/logger"
	"collection-center/library/constant"
	"collection-center/library/redis"
	"collection-center/library/response"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// 限流中间件
func RateLimit(maxCount int64, durationSecond int, process string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.Config().Api.Debug {
			c.Next()
			return
		}
		clientIP := c.ClientIP()
		if clientIP == "::1" {
			clientIP = "127.0.0.1"
		}
		err := checkRequestLimits(clientIP, maxCount, durationSecond, process)
		if err != nil {
			language := c.GetHeader("Language")
			c.AbortWithStatusJSON(http.StatusOK, response.ResErr(err, language))
			return
		}
		c.Next()
	}
}

func checkRequestLimits(ip string, maxCount int64, durationSecond int, process string) error {
	key := constant.RequestLimitKey(ip, process)
	count, err := redis.Client().Incr(key).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		_, err := redis.Client().Expire(key, time.Duration(durationSecond)*time.Second).Result()
		if err != nil {
			logger.Error(err)
		}
	}
	if count > maxCount {
		return ecode.RequestTooFast
	}
	return nil
}
