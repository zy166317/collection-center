package queue

import (
	"collection-center/config"
	"collection-center/internal/logger"
	"collection-center/internal/rpc"
	"collection-center/library/redis"
	"context"
	"encoding/json"
	"github.com/adjust/rmq/v5"
	"strconv"
	"time"
)

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
	verifyOrder := VerifyOrder{}
	if err := json.Unmarshal(payloadByte, verifyOrder); err != nil {
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
		logger.Error("tonRpc.GetTonTransaction:", err)
		_ = delivery.Reject()
		return
	}
	//交易未成功
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
	//if _, has := config.CollectionInfo["TON"]; has {
	//	if transaction.OutMsgs[0].Destination.Value.Address != config.CollectionInfo["TON"].Address {
	//		logger.Info("transaction outmsg dest is not collection address", transaction.Hash)
	//		return
	//	}
	//}

	//通过收款地址校验，插入记录
	//tonRecord := &dao.TonRecord{
	//	Hash:      verifyOrder.Hash,
	//	From:      transaction.Account.Address,
	//	To:        transaction.OutMsgs[0].Destination.Value.Address,
	//	Value:     "",
	//	Status:    "",
	//	CreatedAt: time.Time{},
	//	UpdatedAt: time.Time{},
	//}

	//构建回调信息
	//tonVerifyOrderNotify := &VerifyOrder{
	//	Hash:   transaction.Hash,
	//	Amount: decimal.New(transaction.OutMsgs[0].Value, 10),
	//	IPAddr: verifyOrder.IPAddr,
	//}
	//更新数据库
	//succ, err := dao.UpdateOrderInfo(verifyOrder.Id, &dao.Order{Status: dao.ORDER_SUCCESS, OriginalTokenAmount: tonVerifyOrderNotify.Amount, GameOrderId: tonVerifyOrderNotify.Text})
	//if err != nil || !succ {
	//	logger.Error("UpdateOrderInfo:", err)
	//	_ = delivery.Reject()
	//	return
	//}
	//SendMsg(tonVerifyOrderNotify)
	logger.Info("ton success")
	_ = delivery.Ack()
}
