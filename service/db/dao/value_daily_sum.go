package dao

import (
	"collection-center/service/db"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

// ValueDailySum 每日订单金额表
type ValueDailySum struct {
	ID      int64     `json:"id"`
	UsdtSum string    `json:"usdt_sum"` // 当日兑换U数量, 单位 USDT
	Date    time.Time `json:"date"`     // unique key 日期, 精确到天, 表示该天内 0点 至 24点
}

func (m *ValueDailySum) TableName() string {
	return "value_daily_sum"
}

// InsertValueDailySum 插入每日订单金额
func InsertValueDailySum(data *ValueDailySum) (int64, error) {
	row, err := db.Client().InsertOne(data)
	if err != nil {
		return 0, errors.Wrap(err, "InsertValueDailySum failed")
	}
	if row != 1 {
		return 0, errors.New("Insert failed")
	}
	return data.ID, nil
}

// SelectValueDailySumByDate 根据日期查询每日订单金额
func SelectValueDailySumByDate(date time.Time) (*ValueDailySum, error) {
	data := ValueDailySum{}
	_, err := db.Client().Where("date = ?", date).Get(&data)
	if err != nil {
		return nil, errors.Wrap(err, "SelectValueDailySumByDate failed")
	}
	return &data, nil
}

// SelectValueBetweenDate 查询两个日期之间的订单金额
func SelectValueBetweenDate(startDate, endDate time.Time) ([]ValueDailySum, error) {
	var data []ValueDailySum
	err := db.Client().Where("date >= ?", startDate).And("date <= ?", endDate).Find(&data)
	if err != nil {
		return nil, errors.Wrap(err, "SelectValueBetweenDate failed")
	}
	return data, nil
}

// SumAllValueDailySum 查询所有订单金额
func SumAllValueDailySum() (string, error) {
	var data []ValueDailySum
	err := db.Client().Find(&data)
	if err != nil {
		return "", errors.Wrap(err, "SumAllValueDailySum failed")
	}
	var sum float64
	for _, v := range data {
		f, err := strconv.ParseFloat(v.UsdtSum, 64)
		if err != nil {
			return "", errors.Wrap(err, "SumAllValueDailySum failed")
		}
		sum += f
	}
	return strconv.FormatFloat(sum, 'f', 6, 64), nil
}
