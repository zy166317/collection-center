package service

import (
	"collection-center/internal/ecode"
	"collection-center/internal/logger"
	"collection-center/library/constant"
	"collection-center/library/redis"
	"collection-center/library/utils"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type jwtCustomClaims struct {
	jwt.StandardClaims
	Uid int64 `json:"uid"`
}

func UserLogout(uid int64) error {
	key := constant.AdminTokenKey(uid)
	return redis.Client().Del(key).Err()
}

func AuthByToken(token string) (uid int64, err error) {
	//jwt中解析非敏感信息
	claims, err := ParseToken(token, []byte(constant.JwtSecret))
	if nil != err {
		logger.Warnf("UserAuthByAk error:", err)
		return 0, ecode.NoLogin
	}
	//获取redis token
	Uid := claims.(jwt.MapClaims)["uid"].(float64)
	uid = int64(Uid)
	key := constant.AdminTokenKey(uid)
	redisToken, err := redis.Client().Get(key).Result()
	if err == nil {
		//比对token
		if redisToken != token {
			logger.Warnf("AuthByToken :token not exists")
			return 0, ecode.NoLogin
		}
	} else {
		logger.Warnf("UserAuthByAk error:", err)
		err = ecode.NoLogin
	}
	return
}

func ParseToken(tokenSrt string, SecretKey []byte) (claims jwt.Claims, err error) {
	var token *jwt.Token
	token, err = jwt.Parse(tokenSrt, func(*jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})
	if token != nil {
		claims = token.Claims
	}
	return
}

func CheckPermission(role string, permissionCheck ...string) (hasPer bool) {
	rolePermissionKey := constant.RolePermissionKey(role)
	//获取此角色的权限
	permissionsStr, _ := redis.Client().Get(rolePermissionKey).Result()
	if permissionsStr == "" {
		hasPer = false
	} else {
		var permissions []string
		err := json.Unmarshal([]byte(permissionsStr), &permissions)
		if err != nil {
			return false
		}
		if permissions == nil {
			hasPer = false
		} else {
			for _, perm := range permissionCheck {
				index := utils.FindInString(perm, &permissions)
				if index > -1 {
					hasPer = true
					break
				}
			}
		}
	}
	return
}

func GenerateToken(mail string, Uid int64) (tokenString string, err error) {
	claims := &jwtCustomClaims{
		StandardClaims: jwt.StandardClaims{ExpiresAt: time.Now().Add(time.Hour * constant.AdminExpireHour).Unix(), Issuer: mail},
		Uid:            Uid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	SecretKey := []byte(constant.JwtSecret)
	tokenString, err = token.SignedString(SecretKey)
	return
}

func SaveToken(expireSeconds int, issuer string, Uid int64) (tokenString string, err error) {
	token, err := GenerateToken(issuer, Uid)
	key := constant.AdminTokenKey(Uid)
	if err != nil {
		return token, err
	}
	//存入token
	_, err = redis.Client().Set(key, token, time.Second*time.Duration(expireSeconds)).Result()
	return token, err
}

func SendMailVerifyCodeToRedis(mail string, verifyCode string) (err error) {
	_, err = redis.Client().Set(mail, verifyCode, constant.EmailVerifyCodeExpireMinute*time.Minute).Result()
	return
}

func SaveCaptchaToRedis(idCode string, captCha string) (err error) {
	sendVerifyCodeKey := constant.CaptchaKey(idCode)
	_, err = redis.Client().Set(sendVerifyCodeKey, captCha, constant.CaptchaVerifyExpireMin*time.Minute).Result()
	return
}

func GetCaptchaFromRedis(idCode string) (captcha string, err error) {
	getCaptchaKey := constant.CaptchaKey(idCode)
	captcha, err = redis.Client().Get(getCaptchaKey).Result()
	return
}

func RemoveCaptchaFromRedis(idCode string) (change int64, err error) {
	getCaptchaKey := constant.CaptchaKey(idCode)
	change, err = redis.Client().Del(getCaptchaKey).Result()
	return
}

func GetEmailVerifyCodeFromRedis(mail string) (string, error) {
	result, err := redis.Client().Get(mail).Result()
	return result, err
}
