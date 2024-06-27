package queue

import "github.com/shopspring/decimal"

// 队列查询格式
type VerifyOrder struct {
	OrderId int64
	Hash    string
	Amount  decimal.Decimal
	Text    string
	IPAddr  string
}
