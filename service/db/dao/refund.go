package dao

import (
	"collection-center/service/db"
	"golang.org/x/xerrors"
	"time"
)

type Refund struct {
	Id                   int64     `json:"id" xorm:"pk autoincr not null bigint 'id'"`
	OrderId              int64     `json:"order_id"`
	ReceiveAddr          string    `json:"receive_addr"`
	PlatformReceivedAddr string    `json:"platform_received_addr"`
	Email                string    `json:"email"`
	RefundAmount         string    `json:"refund_amount"`
	RefundToken          string    `json:"refund_token"`
	RefundAddr           string    `json:"refund_addr"`
	CreatedAt            time.Time `json:"created_at" xorm:"created" form:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" xorm:"updated" form:"updated_at"`
}

func InsertRefund(data *Refund) (int64, error) {
	row, err := db.Client().InsertOne(data)
	if err != nil {
		return 0, err
	}
	if row != 1 {
		return 0, xerrors.New("Insert fund failed")
	}
	return data.Id, nil
}

func UpdateRefund(id int64, whereField string, whereVal string, data *Refund) (bool, error) {
	row, err := db.Client().Where("id = ?", id).And(whereField, whereVal).Update(data)
	if err != nil {
		return false, err
	}
	if row != 1 {
		return false, xerrors.New("Update fund failed")
	}

	return true, nil
}

func SelectRefundByOrderID(orderId int64) (*Refund, error) {
	data := Refund{}
	_, err := db.Client().Where("order_id = ?", orderId).Get(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func SelectRefundByLimit(page int, pageSize int) ([]Refund, error) {
	var dataArr []Refund
	err := db.Client().Limit(pageSize, (page-1)*pageSize).Find(&dataArr)
	if err != nil {
		return dataArr, err
	}
	return dataArr, nil
}
