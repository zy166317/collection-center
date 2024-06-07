package queue

import (
	"collection-center/config"
	"collection-center/internal/logger"
	"collection-center/internal/rpc"
	"collection-center/library/redis"
	"collection-center/service"
	"collection-center/service/db/dao"
	"context"
	"encoding/json"
	"github.com/adjust/rmq/v5"
	"github.com/shopspring/decimal"
	"strconv"
	"time"
)

type VerifyOrderNotify struct {
	Hash   string
	Amount decimal.Decimal
	Text   string
	IPAddr string
}

type Comment struct {
	Text string `json:"text"`
}

type TonConsumer struct {
	rmq.Consumer
}

func TonQueueConsumer() error {
	logger.Info("Ton Queue Consumer Starting")
	prefetchLimit := config.Config().Api.QueuePrefetchLimit
	if prefetchLimit == 0 {
		prefetchLimit = 20
	}
	err := redis.TonQueue.StartConsuming(int64(prefetchLimit), 15*time.Second)
	if err != nil {
		logger.Error("FirstQueueConsumer err:", err)
		return err
	}

	// 创建多个Consumer
	for i := 0; i < COSUMER_LIMIT; i++ {
		ret, err := redis.TonQueue.AddConsumer("TonQueueConsumer"+strconv.Itoa(i), new(TonConsumer))
		logger.Infof("第%d个consumer:%s添加成功", i, ret)
		if err != nil {
			logger.Errorf("TonQueueConsumer err:%s", err)
			return err
		}
	}

	//// 创建单个Consumer
	//ret, err := redis.FirstQueue.AddConsumer("FirstQueueConsumer_1", new(FirstConsumer))
	//logger.Infof("consumer:%s添加成功", ret)
	//if err != nil {
	//	logger.Error("FirstQueueConsumer err:", err)
	//	return err
	//}

	go func() {
		// push back the reject message time by time
		for {
			_, err := redis.TonQueue.ReturnRejected(RETURN_LIMIT)
			if err != nil {
				logger.Error(err)
				return
			}
			//logger.Warn("FirstQueue 恢复 ", rejected, " 条 reject 消息")
			time.Sleep(20 * time.Second)
		}
	}()

	logger.Info("TonQueueConsumer Ton队列消费者已启动")
	return nil
}

func (consumer *TonConsumer) Consume(delivery rmq.Delivery) {
	payload := delivery.Payload()
	payloadByte := []byte(payload)
	verifyOrder := &service.VerifyOrder{}
	err := json.Unmarshal(payloadByte, verifyOrder)
	if err != nil {
		logger.Error("TonQueueConsumer 消费失败:", err)
		_ = delivery.Reject()
		return
	}
	//基本交易信息校验
	tonRpc, err := rpc.NewTonRpc()
	if err != nil {
		logger.Error("NewTonRpc:", err)
		_ = delivery.Reject()
		return
	}
	transaction, err := tonRpc.GetTonTransaction(context.Background(), verifyOrder.Hash)
	if err != nil {
		logger.Error("GetTonTransaction:", err)
		_ = delivery.Reject()
		return
	}
	if !transaction.Success {
		logger.Info("transaction status failed", transaction.Hash)
		return
	}
	//from to 校验
	if len(transaction.OutMsgs) <= 0 {
		logger.Info("transaction outmsg is empty", transaction.Hash)
		return
	}
	//收款地址校验
	ok := false
	for _, v := range config.CollectionWalletAddr.TonWallet {
		if transaction.OutMsgs[0].Destination.Value.Address == v {
			ok = true
			break
		}
	}
	if !ok {
		logger.Info("transaction outmsg destination is not equal collectaddr", transaction.Hash)
		return
	}
	//附带信息校验
	if transaction.OutMsgs[0].DecodedBody == nil {
		logger.Info("transaction outmsg decodedbody is nil", transaction.Hash)
		return
	}
	//解析TON交易附带订单号
	var TONComment Comment
	err = json.Unmarshal(transaction.OutMsgs[0].DecodedBody, &TONComment)
	if err != nil {
		logger.Error("json.Unmarshal:", err)
		_ = delivery.Reject()
		return
	}
	//构建回调信息
	tonVerifyOrderNotify := &VerifyOrderNotify{
		Hash:   transaction.Hash,
		Amount: decimal.New(transaction.OutMsgs[0].Value, 10),
		Text:   TONComment.Text,
		IPAddr: verifyOrder.IPAddr,
	}
	//更新数据库
	succ, err := dao.UpdateOrderInfo(verifyOrder.Id, &dao.Order{Status: dao.ORDER_RECEIVED, OriginalTokenAmount: tonVerifyOrderNotify.Amount, GameOrderId: tonVerifyOrderNotify.Text})
	if err != nil || !succ {
		logger.Error("UpdateOrderInfo:", err)
		_ = delivery.Reject()
		return
	}
	logger.Info("test success")
	//TODO 构建通用通知http发送确认信息
	SendMsg(tonVerifyOrderNotify)
	logger.Info(tonVerifyOrderNotify)
}
