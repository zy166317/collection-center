package dao

import "time"

// 付款单
type Payment struct {
	Id                  int64     `json:"id" xorm:"pk autoincr not null bigint 'id'"`
	MerchantUid         int64     `json:"merchant_uid" xorm:"not null  comment('商家uid') bigint 'merchant_uid'"`
	ProjectUid          int64     `json:"project_uid" xorm:"not null  comment('项目uid') bigint 'project_uid'"`
	CollectUid          int64     `json:"collect_uid" xorm:"not null  comment('收款信息uid') bigint 'collect_uid'"`
	CreationChain       string    `json:"creation_chain" xorm:"not null  comment('创建付款单链类型') varchar(64) 'creation_chain'"`
	CreationTokenSymbol string    `json:"creation_token_symbol" xorm:"not null  comment('创建付款单代币简写') varchar(64) 'creation_token_symbol'"`
	CreationAmount      string    `json:"creation_amount" xorm:"not null  comment('创建付款单金额') varchar(255) 'creation_amount'"`
	CreationRate        string    `json:"creation_rate" xorm:"not null  comment('创建付款单时汇率') varchar(255) 'creation_rate'"`
	CreationUValue      string    `json:"creation_u_value" xorm:"not null  comment('创建付款单u值') varchar(255) 'creation_u_value'"`
	ReturnUrl           string    `json:"return_url" xorm:"not null  comment('回调地址') varchar(255) 'return_url'"`
	ExpireTime          int64     `json:"expire_time" xorm:"not null  comment('过期时间') bigint 'expire_time'"`
	PaymentHash         string    `json:"payment_hash" xorm:" comment('付款单hash') varchar(255) 'payment_hash'"` //付款后更新
	FeeHash             string    `json:"fee_hash" xorm:" comment('手续费hash') varchar(255) 'fee_hash'"`
	PaymentStatus       string    `json:"payment_status" xorm:"not null default 'pending' comment('付款单状态') varchar(64) 'payment_status'"`
	FeeStatus           string    `json:"fee_status" xorm:"not null default 'pending' comment('手续费状态') varchar(64) 'fee_status'"`
	PaymentChain        string    `json:"payment_chain" xorm:"  comment('付款链类型') varchar(64) 'payment_chain'"`
	PaymentTokenSymbol  string    `json:"payment_token_symbol" xorm:"comment('付款代币简写') varchar(64) 'payment_token_symbol'"`
	PaymentAmount       string    `json:"payment_amount" xorm:" comment('付款金额') varchar(255) 'payment_amount'"`
	PaymentRate         string    `json:"payment_rate" xorm:"comment('付款时汇率') varchar(255) 'payment_rate'"`
	PaymentUValue       string    `json:"payment_u_value" xorm:"comment('付款u值') varchar(255) 'payment_u_value'"`
	CreatedAt           time.Time `json:"created_at" xorm:"created" form:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" xorm:"updated" form:"updated_at"`
	DeletedAt           time.Time `json:"deleted_at" xorm:"deleted" form:"deleted_at"`
}

func (p *Payment) TableName() string {
	return "payment"
}
