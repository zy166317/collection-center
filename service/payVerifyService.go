package service

import (
	"collection-center/internal/logger"
	"collection-center/library/constant"
	"collection-center/library/redis"
	"collection-center/library/request"
	"collection-center/service/db/dao"
	"encoding/json"
	"errors"
	"time"
)

type VerifyOrder struct {
	Id     int64  //数据库订单ID
	Hash   string //订单hash
	IPAddr string
}

func (v *VerifyOrder) MsgToMarshal() ([]byte, error) {
	marshal, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return marshal, nil
}

func AddPendingTonOrder(req *request.PendingOrderReq) error {
	if req.MsgHash == "" {
		//无效订单
		return errors.New("request params  is invalid")
	}
	//创建订单到数据库
	orderId, err := dao.CreateOrder(&dao.Order{
		OriginalToken: constant.CoinTon,
		Deadline:      time.Now().Add(time.Minute * 30),
		Hash:          req.MsgHash,
		Status:        dao.ORDER_PENDING,
	})
	if err != nil {
		logger.Error(err)
		return err
	}
	//创建Ton待确认订单
	order := &VerifyOrder{
		Id:     orderId,
		Hash:   req.MsgHash,
		IPAddr: req.NotifyIp,
	}
	marshal, err := order.MsgToMarshal()
	if err != nil {
		logger.Error(err)
		return err
	}
	switch req.TokenType {
	case constant.CoinTon:
		err = redis.TonQueue.PublishBytes(marshal)
		if err != nil {
			logger.Error(err)
			return err
		}
	case constant.CoinEth:
		err = redis.ETHQueue.PublishBytes(marshal)
		if err != nil {
			logger.Error(err)
			return err
		}
	case constant.CoinUsdt:
		err = redis.TonQueue.PublishBytes(marshal)
		if err != nil {
			logger.Error(err)
			return err
		}
	case constant.CoinSol:
		err = redis.TonQueue.PublishBytes(marshal)
		if err != nil {
			logger.Error(err)
			return err
		}
	default:
		return errors.New("token type is invalid")
	}
	return nil
}
