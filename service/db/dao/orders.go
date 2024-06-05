package dao

import (
	"collection-center/service/db"
	"encoding/json"
	"github.com/go-xorm/xorm"
	"golang.org/x/xerrors"
	"time"
)

type Orders struct {
	Id                  int64     `json:"id" xorm:"pk autoincr not null bigint 'id'"`
	Status              string    `json:"status" xorm:"not null default '' comment('PENDING, RECEIVED已接收到用户转账;SENDING 给用户转账中; COMPLETED 完成交易 EXPIRED 已过期 REFUND 退款') varchar(9) 'status'"`
	ReceivedTxInfo      string    `json:"received_tx_info" xorm:"default '' comment('用户转账tx详情') varchar(300) 'received_tx_info'"`
	ClosedTxInfo        string    `json:"closed_tx_info" xorm:"default '' comment('核心钱包转账详情') varchar(300) 'closed_tx_info'"`
	Mode                string    `json:"mode" xorm:"not null default '' comment('FIXED 固定费率 / FLOAT 浮动费率') varchar(5) 'mode'"`
	UserReceiveAddress  string    `json:"user_receive_address" xorm:"not null default '' comment('用户接收转账地址') varchar(42) 'user_receive_address'"`
	OriginalToken       string    `json:"original_token" xorm:"not null default '' comment('原始代币, 币种, USDT, BTC, ETH') varchar(10) 'original_token'"`
	OriginalTokenAmount string    `json:"original_token_amount" xorm:"not null default '' comment('原始代币数量') varchar(255) 'original_token_amount'"`
	OriginalTokenToU    string    `json:"original_token_to_u" xorm:"original_token_to_u"`
	TargetToken         string    `json:"target_token" xorm:"not null default '' comment('目标代币, 币种, USDT, BTC, ETH') varchar(10) 'target_token'"`
	TargetTokenAmount   string    `json:"target_token_amount" xorm:"not null default '' comment('目标代币数量') varchar(255) 'target_token_amount'"`
	TargetTokenReceived string    `json:"target_token_received"`
	WeReceiveAddress    string    `json:"we_receive_address" xorm:"not null default '' comment('平台收币地址') varchar(42) 'we_receive_address'"`
	Email               string    `json:"email" xorm:"default '' comment('用户邮箱地址, 可空') varchar(100) 'email'"`
	Deadline            time.Time `json:"deadline" xorm:"not null comment('30分钟到期时间') datetime(6) 'deadline'"`
	CreatedAt           time.Time `json:"created_at" xorm:"created" form:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" xorm:"updated" form:"updated_at"`
}

func (o *Orders) TableName() string {
	return "orders"
}

// order status 说明
// PENDING: 未接收到用户转账
// RECEIVED: 已接收到用户转账
// SENDING: 给用户转账中
// COMPLETED: 完成交易
// EXPIRED: 已过期
// REFUND: 退款

/*
	PENDING ==> EXPIRED ==> REFUND
	  ||
	  ||
	  \/
	RECEIVED  已接收到用户转账
	  ||
	  ||
	  \/
	SENDING  给用户转账中 ==> ERROR_SENDING 手动处理的状态
	  ||
	  ||
	  \/
	COMPLETED  完成交易

*/

const (
	// ORDER_PENDING PENDING 未接收到用户转账
	ORDER_PENDING = "PENDING"
	// ORDER_RECEIVED RECEIVED 已接收到用户转账
	ORDER_RECEIVED = "RECEIVED"
	// ORDER_SENDING SENDING 给用户转账中
	ORDER_SENDING = "SENDING"
	// ORDER_COMPLETED COMPLETE 完成交易
	ORDER_COMPLETED = "COMPLETED"
	// ORDER_EXPIRED EXPIRED 已过期
	ORDER_EXPIRED = "EXPIRED"
	// ORDER_REFUND REFUND 退款
	ORDER_REFUND = "REFUND"
)

func (o *Orders) MarshalToMsg() ([]byte, error) {
	bs, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func InsertOrder(data *Orders) (int64, error) {
	row, err := db.Client().InsertOne(data)
	if err != nil {
		return 0, err
	}
	if row != 1 {
		return 0, xerrors.New("Insert failed")
	}
	return data.Id, nil
}

func UpdateOrder(id int64, curStatus string, order *Orders) (bool, error) {
	row, err := db.Client().Where("id = ?", id).And("status = ?", curStatus).Update(order)
	if err != nil {
		return false, err
	}
	if row != 1 {
		return false, xerrors.New("Update failed")
	}

	return true, nil
}

func UpdateOrderStatusById(session *xorm.Session, id int64, curStatus, nextStatus string) error {
	row, err := session.ID(id).And("status = ?", curStatus).Cols("status").Update(&Orders{Status: nextStatus})
	if err != nil {
		return err
	}
	if row != 1 {
		return xerrors.New("Update failed")
	}

	return nil
}

func SelectOrderByID(id int64) (*Orders, error) {
	order := Orders{}
	_, err := db.Client().Where("id = ?", id).Get(&order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func SelectOrdersByLimit(page int, pageSize int) ([]Orders, error) {
	var orders []Orders
	err := db.Client().Limit(pageSize, (page-1)*pageSize).Find(&orders)
	if err != nil {
		return orders, err
	}
	return orders, nil
}

func SelectOrdersLimitNWhere(page int, pageSize int, whereField string, whereVal string) ([]Orders, error) {
	var orders []Orders
	err := db.Client().Where(whereField, whereVal).Limit(pageSize, (page-1)*pageSize).Desc("id").Find(&orders)
	if err != nil {
		return orders, err
	}
	return orders, nil
}

func SelectOrders(session *xorm.Session, page int, pageSize int) ([]Orders, error) {
	var orders []Orders
	err := session.Limit(pageSize, (page-1)*pageSize).Desc("id").Find(&orders)
	if err != nil {
		return orders, err
	}
	return orders, nil
}

// CountOrderByStatus 统计已完成的订单数量
// gtId 大于 该id 的订单数量 - 传 0 则统计所有已完成的订单数量
func CountOrderByStatus(session *xorm.Session, gtId int64, status string) (int64, error) {
	var orders Orders
	s := session.Where("status = ?", status)
	if gtId > 0 {
		s = s.Where("id > ?", gtId)
	}
	total, err := s.Count(&orders)
	if err != nil {
		return 0, err
	}
	return total, nil
}
