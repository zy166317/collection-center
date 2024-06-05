package constant

import (
	"strconv"
	"strings"
	"time"
)

const RedisAdminTokenKey = "SWAPX.TOKEN"
const TmpPwd = "SWAPX.AUTH.TMP.PWD"
const RedisRolePermissionKey = "SWAPX.PERMISSION"
const STSConfigKey = "SWAPX.STS.CONFIG"
const RedisEmailVerifyKey = "SWAPX.EmailVerify"
const RedisCaptchaVerifyKey = "SWAPX.Captcha"
const RedisRequestKey = "SWAPX.LIMIT.Request"
const RedisFuncKey = "SWAPX.LIMIT.Func"
const JwtSecret = "salt_19!iulia1@!&0"
const AdminExpireHour = 12
const TmpExpireMinutes = 120
const EmailVerifyExpireHour = 2
const CaptchaVerifyExpireMin = 2
const SessionExpireHour = 24 * 7
const SessionValue = "SWAPX:SESSION:VALUE"
const ExcelProcessCountMax = 8
const (
	AcquireLockTimeout = 10 * time.Second
	LockTimeout        = 60 * time.Second
)
const (
	// FirstListenSamWalletQueue 监听子账户队列,存储创建订单后的消息(消息体为 orders marshal)
	FirstListenSamWalletQueue  = "QUEUE:First_Listen_Sam_Wallet"
	SecondListenSamWalletQueue = "QUEUE:Second_Listen_Sam_Wallet"
	ThirdListenSamWalletQueue  = "QUEUE:Third_Listen_Sam_Wallet"
	CoreToUserQueue            = "QUEUE:Core_To_User"
)

const (
	// OrderValueAll 缓存订单价值信息的key, 记录 全部时间 的订单价值 (单位: usdt)
	OrderValueAll = "ORDER:VALUE:ALL"
)

func GetSessionKey(sessionValue string) string {
	return strings.Join([]string{SessionValue, sessionValue}, ":")
}

// GetLqLockedKey 获取 流动性锁仓 key, 单位: BTC/ETH/USDT 例如 0.1BTC锁仓, key为 LQ:LOCKED:BTC, value为 0.1
func GetLqLockedKey(coin string) string {
	return strings.Join([]string{"LQ:LOCKED", coin}, ":")
}

func AdminTokenKey(uid int64) string {
	return strings.Join([]string{RedisAdminTokenKey, strconv.FormatInt(uid, 10)}, ":")
}
func RolePermissionKey(role string) string {
	return strings.Join([]string{RedisRolePermissionKey, role}, ":")
}

func TmpPwdKey(username string) string {
	return strings.Join([]string{TmpPwd, username}, ":")
}

func MailVerifyCodeKey(uid int64) string {
	return strings.Join([]string{RedisEmailVerifyKey, strconv.FormatInt(uid, 10)}, ":")
}

func CaptchaKey(idCode string) string {
	return strings.Join([]string{RedisCaptchaVerifyKey, idCode}, ":")
}

func RequestLimitKey(ip string, process string) string {
	return strings.Join([]string{RedisRequestKey + "." + process, ip}, ":")
}

func FuncLimitKey(funcName, paramsHash string) string {
	return strings.Join([]string{RedisFuncKey + ":" + funcName, paramsHash}, ":")
}
