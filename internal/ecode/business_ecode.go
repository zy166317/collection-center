package ecode

//// blockchain verify error code
//var (
//	InvalidTransactionId = add(20001)
//)

var (
	IllegalParam               = add(10001) //参数错误
	EmailIsExist               = add(10101) //邮箱已存在
	VerifyCodeError            = add(10102) //验证码错误
	EmailPwdNotMatch           = add(10103) //邮箱密码不匹配
	NoLogin                    = add(10104) // 账号未登录
	CollectInfoAddressNotMatch = add(10105) //收款信息与收款地址不匹配
	TokenSymbolNotExist        = add(10106) //token symbol 不存在
	EmailIsNotExist            = add(10107) //邮箱不存在
	UsdtRateNotChange          = add(10108) //USDT汇率不能修改
	RateMustBePositive         = add(10109) //费率必须大于0
	CollectAddressFormatError  = add(10110) //收款地址格式错误
	CreateProjectError         = add(10111) //数据库错误
	CheckTokenAddressError     = add(10112) //检测token地址错误
	AddTokenInfoFailed         = add(10113) //添加token信息失败
	ProjectNotExist            = add(10114) //项目不存在
	NoProjectPermission        = add(10115) //没有项目权限
	UpdateProjectInfoFailed    = add(10116) //更新项目信息失败
	UpdateCollectAddressFailed = add(10117) //更新收款地址失败
	UpdateCollectRateFailed    = add(10118) //更新汇率失败
	FreezeProjectFailed        = add(10119) //冻结项目失败
	RequestTooFast             = add(10207) // 触发频率限制
)
