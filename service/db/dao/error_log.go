package dao

import (
	"collection-center/service/db"
	"time"
)

type ErrorLog struct {
	Id             int64     `json:"id" xorm:"pk autoincr not null bigint 'id'"`
	OrderId        int64     `json:"order_id"`         // 订单id
	CurOrderStatus string    `json:"cur_order_status"` // 错误前订单状态
	ErrorLog       string    `json:"error_log"`        // 错误信息
	ErrorHash      string    `json:"error_hash"`       // 错误交易hash
	ErrorCoin      string    `json:"error_coin"`       // 错误币种 ETH/USDT/BTC
	CreatedAt      time.Time `json:"created_at" xorm:"created"`
}

func (m *ErrorLog) TableName() string {
	return "error_log"
}

func InsertErrorLog(data *ErrorLog) error {
	_, err := db.Client().InsertOne(data)
	if err != nil {
		return err
	}
	return nil
}

func InsertErrorLogByOrder(orderId int64, status, coin, hash string, err error) error {
	data := ErrorLog{
		OrderId:        orderId,
		CurOrderStatus: status,
		ErrorLog:       err.Error(),
		ErrorHash:      hash,
		ErrorCoin:      coin,
		CreatedAt:      time.Now(),
	}
	return InsertErrorLog(&data)
}
