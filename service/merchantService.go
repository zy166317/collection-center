package service

import (
	"collection-center/internal/ecode"
	"collection-center/internal/email"
	"collection-center/library/request"
	"collection-center/library/utils"
	"collection-center/service/db/dao"
	"fmt"
	"math/rand"
	"time"
)

// SendVerifyCode 发送邮箱验证代码。
// 该函数用于在用户注册或修改邮箱时，向指定邮箱发送验证代码。
// req  包含待验证邮箱信息的请求结构体。
// error 如果发送过程中出现错误，则返回错误对象；否则返回nil。
func SendVerifyCode(req *request.SendVerifyCodeReq) error {
	// 检查邮箱是否已存在
	mail, err := dao.IsExistEmail(req.Email)
	if err != nil {
		return err
	}
	if mail == true {
		return ecode.EmailIsExist
	}

	// 生成6位随机验证代码
	code := fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))

	// 将验证代码存储到Redis中，以供后续验证使用
	err = SendMailVerifyCodeToRedis(req.Email, code)
	if err != nil {
		return err
	}

	// 发送验证邮件到指定邮箱
	email.SendEmail(&email.BaseMailInfo{
		MailAddress:      req.Email,
		VerificationCode: code,
	})
	return nil
}

// RegisterMerchant 注册商家账户。
// req: 商家注册请求，包含邮箱、验证码和密码等信息。
// 返回值: 错误信息，如注册过程中出现任何问题。
func RegisterMerchant(req *request.CreateMerchantReq) error {
	// 检查注册请求中的必要参数是否为空
	if req.VerifyCode == "" || req.Email == "" || req.Password == "" {
		return ecode.IllegalParam
	}

	// 从Redis中获取邮箱对应的验证码
	//redis验证验证码
	value, err := GetEmailVerifyCodeFromRedis(req.Email)
	if err != nil {
		return err
	}

	// 验证请求中的验证码与Redis中存储的验证码是否一致
	if req.VerifyCode != value {
		return ecode.VerifyCodeError
	}

	// 创建商家对象，并初始化相关字段
	// 创建商家信息
	merchant := &dao.Merchant{
		Email:               req.Email,
		MerchantUid:         utils.GenerateUid(),
		Password:            req.Password,
		MerchantAuditStatus: dao.MerchantAuditStatusPending,
		MerchantStatus:      dao.MerchantStatusNormal,
	}

	// 将商家对象插入数据库
	_, err = dao.CreateMerchant(merchant)
	if err != nil {
		return err
	}

	// 注册成功，返回nil
	return nil
}

// LoginMerchant 为商家登录接口。
// req: 商家登录请求，包含邮箱、密码信息。
// 返回值: 错误信息，如登录过程中出现任何问题。
func LoginMerchant(req *request.LoginMerchantReq) (string, error) {
	// 验证邮箱和密码是否为空
	if req.Email == "" || req.Password == "" {
		return "", ecode.IllegalParam
	}

	// 通过邮箱获取商家信息
	// 判断账号密码信息
	merchant, err := dao.GetMerchantByEmail(req.Email)
	if err != nil || merchant == nil {
		return "", ecode.EmailIsNotExist
	}

	// 验证密码是否匹配
	if req.Password != merchant.Password {
		return "", ecode.EmailPwdNotMatch
	}

	// 生成并返回token
	// 验证通过生成token令牌
	token, err := SaveToken(3600, req.Email, merchant.MerchantUid)
	return token, nil
}
