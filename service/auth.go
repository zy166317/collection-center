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
	Uid              int64  `json:"uid"`
	Role             string `json:"role"`
	Username         string `json:"username"`
	PlatFormType     string `json:"platformType"`
	Nickname         string `json:"nickname"`
	OrganizationCode string `json:"organizationCode"`
}

func UserLogout(uid int64) error {
	key := constant.AdminTokenKey(uid)
	return redis.Client().Del(key).Err()
}

func AuthByToken(token string) (uid int64, role string, organizationCode string, platformType string, username string, nickname string, err error) {
	//jwt中解析非敏感信息
	claims, err := ParseToken(token, []byte(constant.JwtSecret))
	if nil != err {
		logger.Warnf("UserAuthByAk error:", err)
		return 0, "", "", "", "", "", ecode.NoLogin
	}
	//获取redis token
	Uid := claims.(jwt.MapClaims)["uid"].(float64)
	uid = int64(Uid)
	role = claims.(jwt.MapClaims)["role"].(string)
	platformType = claims.(jwt.MapClaims)["platformType"].(string)
	key := constant.AdminTokenKey(uid)
	username = claims.(jwt.MapClaims)["username"].(string)
	nickname = claims.(jwt.MapClaims)["nickname"].(string)
	redisToken, err := redis.Client().Get(key).Result()
	if err == nil {
		//比对token
		if redisToken != token {
			logger.Warnf("AuthByToken :token not exists")
			return 0, "", "", "", "", "", ecode.NoLogin
		}
		organizationCode = claims.(jwt.MapClaims)["organizationCode"].(string)
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

func GenerateToken(issuer string, Uid int64, Role string, OrganizationCode string, PlatFormType, nickname string) (tokenString string, err error) {
	claims := &jwtCustomClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * constant.AdminExpireHour).Unix(),
			Issuer:    issuer,
		},
		Uid,
		Role,
		// issuer 就是 username
		issuer,
		PlatFormType,
		nickname,
		OrganizationCode,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	SecretKey := []byte(constant.JwtSecret)
	tokenString, err = token.SignedString(SecretKey)
	return
}

func SaveToken(expireSeconds int, issuer string, Uid int64, Role string, OrganizationCode string, PlatFormType, nickname string) (tokenString string, err error) {
	token, err := GenerateToken(issuer, Uid, Role, OrganizationCode, PlatFormType, nickname)
	key := constant.AdminTokenKey(Uid)
	if err != nil {
		return token, err
	}
	//存入token
	_, err = redis.Client().Set(key, token, time.Second*time.Duration(expireSeconds)).Result()
	return token, err
}

func SendMailVerifyCodeToRedis(uid int64, verifyCode string) (err error) {
	sendVerifyCodeKey := constant.MailVerifyCodeKey(uid)
	_, err = redis.Client().Set(sendVerifyCodeKey, verifyCode, constant.EmailVerifyExpireHour*time.Hour).Result()
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
