package request

type CreateUser struct {
	Username         string `json:"username"`
	UserPhone        string `json:"userPhone"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	Role             string `json:"role"`
	Nickname         string `json:"nickname"`
	OrganizationCode string `json:"organizationCode"`
	AccountStatus    string `json:"accountStatus"`
}

type UpdateUser struct {
	UserId        int64  `json:"userId"`
	UserPhone     string `json:"userPhone"`
	Email         string `json:"email"`
	Password      string `json:"password"`
	Role          string `json:"role"`
	Nickname      string `json:"nickname"`
	AccountStatus string `json:"accountStatus"`
}

type UserPassLogin struct {
	VerifyCaptcha
	Password     string `json:"password"`
	Username     string `json:"username"`
	PlatformType string `json:"platformType"`
}

type UpdateUserStatus struct {
	UserId        int64  `json:"userId"`
	AccountStatus string `json:"accountStatus"`
}
type EmailVerify struct {
	Email string `json:"email"`
}

type UpdateActiveStatus struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	EmailVerifyCode string `json:"emailVerifyCode"`
	PlatformType    string `json:"platformType"`
}
type ListUserTmpPwdReq struct {
	UsernameList *[]string `json:"usernameList"`
}
type CreateTmpPwdReq struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	ExpireSeconds int64  `json:"expireSeconds"`
}

type VerifyCaptcha struct {
	IdCode    string `json:"idCode"`
	ImageCode string `json:"imageCode"`
}
