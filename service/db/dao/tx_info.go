package dao

import "time"

// 确认收款到账表
type TxInfo struct {
	Id          int64     `json:"id" xorm:"pk autoincr not null bigint 'id'"`
	PaymentId   int       `json:"payment_id" xorm:"unique not null  comment('付款单id') int 'payment_id'"`
	From        string    `json:"from" xorm:"not null  comment('付款地址') varchar(255) 'from'"`
	To          string    `json:"to" xorm:"not null  comment('收款地址') varchar(255) 'to'"`
	Chain       string    `json:"chain" xorm:"not null  comment('链名') varchar(64) 'chain'"` //链名
	TokenSymbol string    `json:"token_symbol" xorm:"not null  comment('代币简写') varchar(64) 'token_symbol'"`
	PayAmount   string    `json:"pay_amount" xorm:"not null  comment('付款金额') varchar(255) 'pay_amount'"`
	TxHash      string    `json:"tx_hash" xorm:"not null  comment('交易hash') varchar(255) 'tx_hash'"`
	CreatedAt   time.Time `json:"created_at" xorm:"created" form:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" xorm:"updated" form:"updated_at"`
	DeletedAt   time.Time `json:"deleted_at" xorm:"deleted" form:"deleted_at"`
}
