package queue

import (
	"collection-center/config"
	"collection-center/internal/logger"
	"collection-center/internal/rpc"
	"collection-center/library/redis"
	"collection-center/library/utils"
	"collection-center/service"
	"collection-center/service/db/dao"
	"context"
	"encoding/json"
	"github.com/adjust/rmq/v5"
	"github.com/shopspring/decimal"
	"strconv"
	"time"
)

type EthConsumer struct {
	rmq.Consumer
}

func EthQueueConsumer() error {
	logger.Info("Ton Queue Consumer Starting")
	prefetchLimit := config.Config().Api.QueuePrefetchLimit
	if prefetchLimit == 0 {
		prefetchLimit = 20
	}
	err := redis.ETHQueue.StartConsuming(int64(prefetchLimit), 15*time.Second)
	if err != nil {
		logger.Error("FirstQueueConsumer err:", err)
		return err
	}

	// 创建多个Consumer
	for i := 0; i < COSUMER_LIMIT; i++ {
		ret, err := redis.ETHQueue.AddConsumer("EthQueueConsumer"+strconv.Itoa(i), new(EthConsumer))
		logger.Infof("第%d个consumer:%s添加成功", i, ret)
		if err != nil {
			logger.Errorf("EthQueueConsumer err:%s", err)
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
			_, err := redis.ETHQueue.ReturnRejected(RETURN_LIMIT)
			if err != nil {
				logger.Error(err)
				return
			}
			//logger.Warn("FirstQueue 恢复 ", rejected, " 条 reject 消息")
			time.Sleep(20 * time.Second)
		}
	}()

	logger.Info("EthQueueConsumer Ton队列消费者已启动")
	return nil
}

func (consumer *EthConsumer) Consume(delivery rmq.Delivery) {
	payload := delivery.Payload()
	payloadByte := []byte(payload)
	verifyOrder := &service.VerifyOrder{}
	err := json.Unmarshal(payloadByte, verifyOrder)
	if err != nil {
		logger.Error("EthQueueConsumer 消费失败:", err)
		_ = delivery.Reject()
		return
	}
	ethRpc, err := rpc.NewBaseEthRpc(false)
	if err != nil {
		logger.Error("NewTonRpc:", err)
		_ = delivery.Reject()
		return
	}
	transaction, err := ethRpc.GetTransactionByTxSign(context.Background(), verifyOrder.Hash)
	if err != nil {
		logger.Error("GetTransferByTxSign:", err)
		_ = delivery.Reject()
		return
	}
	if transaction.Input == "0x" {
		logger.Error("transaction input is null:", verifyOrder.Hash)
		return
	}
	ok := false
	for _, v := range config.CollectionWalletAddr.EthWallet {
		if v == transaction.To {
			ok = true
			break
		}
	}
	if !ok {
		logger.Error("transaction to is not equal payAddr:", verifyOrder.Hash)
		return
	}
	//交易订单校验
	if transaction.Input == "" {
		logger.Error("transaction input is null:", verifyOrder.Hash)
		return
	}
	//amount格式转换
	amount, err := decimal.NewFromString(transaction.Value)
	if err != nil {
		logger.Error("NewFromString:", err)
		_ = delivery.Reject()
		return
	}
	//eth input16进制转string
	text, err := utils.HexadecimalToString(transaction.Input)
	if err != nil {
		logger.Error("HexadecimalToString:", err)
		_ = delivery.Reject()
		return
	}
	//构建回调信息
	tonVerifyOrderNotify := &VerifyOrderNotify{
		Hash:   transaction.Hash,
		Amount: amount,
		Text:   text,
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
	logger.Info(tonVerifyOrderNotify)
}
