package dao

import (
	"collection-center/service/db"
	"github.com/shopspring/decimal"
	"golang.org/x/xerrors"
	"time"
)

type Order struct {
	Id                  int64           `json:"id" xorm:"pk autoincr not null bigint 'id'"`
	Hash                string          `json:"hash" xorm:"not null default '' comment('订单hash') varchar(255) 'hash'"`
	Status              string          `json:"status" xorm:"not null default '' comment('订单状态 SUCCESS,FAILED') varchar(255) 'status'"`
	CollectionAddress   string          `json:"collectionAddress" xorm:"not null default '' comment('收款地址') varchar(255) 'collectionAddress'"`
	OriginalToken       string          `json:"originalToken" xorm:"not null default '' comment('原始代币, 币种, USDT, SOL, ETH,TON') varchar(10) 'originalToken'"`
	OriginalTokenAmount decimal.Decimal `json:"originalTokenAmount" xorm:"not null default '' comment('原始代币数量') decimal(20,8) 'originalTokenAmount'"`
	GameOrderId         string          `json:"gameOrderId" xorm:"not null default '' comment('游戏订单id') varchar(255) 'gameOrderId'"`
	Deadline            time.Time       `json:"deadline" xorm:"not null comment('30分钟到期时间') datetime(6) 'deadline'"`
	CreatedAt           time.Time       `json:"created_at" xorm:"created" form:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at" xorm:"updated" form:"updated_at"`
}

func (o *Order) TableName() string {
	return "order"
}

//const (
//	ORDER_PENDING = "PENDING"
//	ORDER_SUCCESS = "SUCCESS"
//	ORDER_FAILED  = "FAILED"
//)

func CreateOrder(data *Order) (int64, error) {
	//根据hash判断当前订单是否存在
	exist, err := db.Client().Where("hash = ?", data.Hash).Exist(new(Order))
	if err != nil {
		return 0, err
	}
	if exist {
		return 0, xerrors.New("Order already exists")
	}
	row, err := db.Client().InsertOne(data)
	if err != nil {
		return 0, err
	}
	if row != 1 {
		return 0, xerrors.New("Create failed")
	}
	return data.Id, nil
}

func UpdateOrderInfo(id int64, order *Order) (bool, error) {
	row, err := db.Client().Where("id = ?", id).Update(order)
	if err != nil {
		return false, err
	}
	if row != 1 {
		return false, xerrors.New("Update failed")
	}

	return true, nil
}
