package request

// 发送验证码
type SendVerifyCodeReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateMerchantReq struct {
	VerifyCode string `json:"verifyCode"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

type LoginMerchantReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateProjectReq struct {
	Name           string                  `json:"name"`      //项目名称
	Domain         string                  `json:"domain"`    //项目域名
	NotifyUrl      string                  `json:"notifyUrl"` //项目回调地址
	CollectInfo    map[string][]*TokenInfo //收款信息
	CollectAddress map[string]string       //收款地址
}

type TokenInfo struct {
	TokenSymbol string `json:"tokenSymbol"` //币种
	Rate        int    `json:"rate"`        //汇率必须为整数
}

type AddTokenInfoReq struct {
	Chain           string `json:"chain"`
	ContractAddress string `json:"contractAddress"`
	LogoUrl         string `json:"logoUrl"`
}

type UpdateProjectInfo struct {
	ProjectUid int64  `json:"projectUid"`
	Domain     string `json:"domain"`
	NotifyUrl  string `json:"notifyUrl"`
}

// 更新收款信息
type UpdateCollectRate struct {
	ProjectUid int64 `json:"projectUid"`
	CollectUid int64 `json:"collectUid"`
	Rate       int   `json:"rate"`
}

// 更新收款地址
type UpdateCollectAddress struct {
	ProjectUid int64  `json:"projectUid"`
	Chain      string `json:"chain"`
	Address    string `json:"address"`
}

// 冻结项目
type FreezeProjectReq struct {
	ProjectUid int64 `json:"projectUid"`
}

// 创建付款单
type CreatePaymentReq struct {
	ProjectUid          int64  `json:"projectUid"`
	CollectUid          int64  `json:"collectUid"`
	CreationChain       string `json:"creationChain"`
	CreationTokenSymbol string `json:"creationTokenSymbol"`
	ReturnUrl           string `json:"returnUrl"`
}
