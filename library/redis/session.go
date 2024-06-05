package redis

import (
	"collection-center/internal/logger"
	"collection-center/library/constant"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"net/http"
	"strings"
)

var sessionStore redis.Store

// initRedisSession 初始化 session 存储
func initRedisSession(size int, network, address, password, DB string, domain string, keyPairs ...[]byte) (redis.Store, error) {
	logger.Info("=============InitRedisSession start")

	store, err := redis.NewStoreWithDB(size, network, address, password, DB, keyPairs...)
	if err != nil {
		logger.Error("NewStore err:", err)
		return nil, err
	}
	ops := &sessions.Options{
		MaxAge:   constant.SessionExpireHour * 60 * 60,
		Path:     "/",
		HttpOnly: true,
		//Secure:   false,
		//SameSite: http.SameSiteNoneMode,
	}
	if strings.Contains(domain, "https") {
		ops.Secure = true
		ops.SameSite = http.SameSiteStrictMode
	}
	store.Options(*ops)

	logger.Info("=============InitRedisSession end")
	return store, nil
}

func GetSessionStore() redis.Store {
	return sessionStore
}
