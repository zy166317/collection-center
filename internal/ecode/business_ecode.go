package ecode

var (
	IllegalParam        = add(10001) //参数错误
	InvalidClientInfo   = add(11001)
	InvalidAccessToken  = add(11002)
	InvalidRefreshToken = add(11003)

	LoginFailed           = add(10101) //登陆失败
	AccessTokenExpires    = add(10102) // Token 过期
	NoLogin               = add(10103) // 账号未登录
	CaptchaErr            = add(10104) // 验证码错误
	FailedTooManyTimes    = add(10105) // 登录失败次数太多
	UserNotExist          = add(10106) // 用户不存在
	UsernameOrPasswordErr = add(10107) // 用户名或密码错误
	InvalidToken          = add(10108) // token 非法
	LoginTooFast          = add(10109) // 触发登录限制
	SendSmsTooFast        = add(10110) // 验证码发送过于频繁，请60s后重试
	SendSmsIllegalRegion  = add(10111) // 目前短信发送尚不支持此地区
	InvalidEmailSender    = add(10130) // 不存在的发送邮箱

	UserBlocked          = New(10201) // 用户被封禁
	UserNameDuplicate    = add(10202) // 重复的用户
	UserNameFormatErr    = New(10203) // 用户名不合法
	MobileFormatErr      = New(10204) // 手机号不合法
	PasswordTooLeak      = add(10205) // 密码太弱
	UserNameOverLimit    = New(10206) // 用户名长度超过限制
	RequestTooFast       = add(10207) // 触发频率限制
	SamePasswordInModify = add(10208) // 修改密码时， 新密码与旧密码相同
	InvalidInviteCode    = add(10209) // 无效的邀请码
	AlreadyKycIdCard     = add(10210) // 已验证过身份证，不可重复提交
	IllegalIdCard        = add(10211) // 身份证号不合法
	IllegalIdRole        = add(10212) // 角色不合法
	IllegalIdRoleChange  = add(10213) // 角色不允许跨平台便跟
	EmailVerifyCodeError = add(10214) // 邮箱验证码错误

	BlockchainUploadError = add(20001) // 区块链上传错误
	BlockchainError       = add(20002) // 区块链错误
	InfoAlreadyOnChain    = add(21001) //信息已生效无法修改

)
