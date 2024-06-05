package dao

import (
	"github.com/go-xorm/xorm"
	"time"
)

type Steps struct {
	Id             int64     `xorm:"pk autoincr not null bigint 'id'"`
	Status         string    `xorm:"not null default '' comment('0-4, 0-订单生成了, 1-成功拿到收款tx数据, 2-只有USDT打Gas费有此状态, 3-归集成功, 4-打款, 5-打款tx') varchar(32) 'status'"`
	OrderId        int64     `xorm:"not null bigint 'order_id'"`
	ReceivedTxHash string    `xorm:"default '' varchar(300) 'received_tx_hash'"`
	ClosedTxHash   string    `xorm:"default '' varchar(300) 'closed_tx_hash'"`
	CreatedAt      time.Time `xorm:"default 'CURRENT_TIMESTAMP' timestamp 'created_at'"`
	UpdatedAt      time.Time `xorm:"default 'CURRENT_TIMESTAMP' timestamp 'updated_at'"`
}

const (
	StepsStatusCollect = "3" // 归集成功
)

func (s *Steps) TableName() string {
	return "steps"
}

func InsertCollectStep(session *xorm.Session, orderId int64, receivedTxHash string) error {
	steps := Steps{
		Status:         StepsStatusCollect,
		OrderId:        orderId,
		ReceivedTxHash: receivedTxHash,
		CreatedAt:      time.Now(),
	}

	_, err := session.Insert(&steps)
	return err
}
