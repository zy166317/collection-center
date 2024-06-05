package redis

import (
	"collection-center/internal/logger"
	"collection-center/library/constant"
	"time"
)

// RateLimitForFunc 函数限流器
func RateLimitForFunc(funcName, paramsHash string, expireSecond int) bool {
	// 获取 funcName:paramsHash 是否存在
	key := constant.FuncLimitKey(funcName, paramsHash)
	if Client().Exists(key) {
		return false
	}
	// 写入 redis
	_, err := Client().Set(key, 1, time.Duration(expireSecond)*time.Second).Result()
	if err != nil {
		logger.Error(err)
		return false
	}
	return true
}
