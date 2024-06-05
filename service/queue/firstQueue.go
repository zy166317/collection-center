package queue

import (
	"collection-center/config"
	cnt "collection-center/contract/constant"
	"collection-center/internal/btc"
	"collection-center/internal/logger"
	"collection-center/internal/rpc"
	"collection-center/library/redis"
	"collection-center/service"
	"collection-center/service/db/dao"
	"encoding/json"
	"fmt"
	"github.com/adjust/rmq/v5"
	"strconv"
	"time"
)

type FirstConsumer struct {
	rmq.Consumer
}

// FirstQueueConsumer 第一队列消费者-监听子账户消费者
func FirstQueueConsumer() error {
	logger.Info("FirstQueueConsumer StartConsuming 1")
	prefetchLimit := config.Config().Api.QueuePrefetchLimit
	if prefetchLimit == 0 {
		prefetchLimit = 20
	}
	err := redis.FirstQueue.StartConsuming(int64(prefetchLimit), 15*time.Second)
	if err != nil {
		logger.Error("FirstQueueConsumer err:", err)
		return err
	}

	// 创建多个Consumer
	for i := 0; i < COSUMER_LIMIT; i++ {
		ret, err := redis.FirstQueue.AddConsumer("FirstQueueConsumer_"+strconv.Itoa(i), new(FirstConsumer))
		logger.Infof("第%d个consumer:%s添加成功", i, ret)
		if err != nil {
			logger.Errorf("FirstQueueConsumer err:%s", err)
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
			_, err := redis.FirstQueue.ReturnRejected(RETURN_LIMIT)
			if err != nil {
				logger.Error(err)
				return
			}
			//logger.Warn("FirstQueue 恢复 ", rejected, " 条 reject 消息")
			time.Sleep(20 * time.Second)
		}
	}()

	logger.Info("FirstQueueConsumer 第一队列消费者已启动")
	return nil
}

func (consumer *FirstConsumer) Consume(delivery rmq.Delivery) {
	// 提取队列数据
	payload := delivery.Payload()
	payloadByte := []byte(payload)

	var hashOrder service.HashOrder
	err := json.Unmarshal(payloadByte, &hashOrder)
	if err != nil {
		logger.Error("[First Queue] Order 消息处理失败, 重新推回队列", err)
		_ = delivery.Reject()
		return
	}

	order := hashOrder.Order

	// 检查是否到了 InitHeight, 没到则重新推回队列
	if order.OriginalToken == "ETH" || order.OriginalToken == "USDT" {
		nowBlock, err := redis.GetHeightFormRedis(cnt.ETH_HEIGHT)
		if err != nil {
			logger.Error(err)
			_ = delivery.Reject()
			return
		}

		waitHeight := hashOrder.InitHeight + (uint64(rpc.WaitBlock) * 2)

		if nowBlock < waitHeight {
			_ = delivery.Reject()
			return
		}
	} else if order.OriginalToken == "BTC" {
		nowBlock, err := redis.GetHeightFormRedis(cnt.BTC_HEIGHT)
		if err != nil {
			logger.Error(err)
			_ = delivery.Reject()
			return
		}

		waitHeight := hashOrder.InitHeight + uint64(btc.WaitBlock)
		if nowBlock < waitHeight {
			_ = delivery.Reject()
			return
		}
	} else {
		logger.Error("Invalid order")
		_ = delivery.Ack()
		return
	}

	logger.Infof(
		"[First Queue] Order[%d][%v To %v] is processing",
		order.Id,
		order.OriginalToken,
		order.TargetToken,
	)

	now := time.Now()
	// 检测订单是否超时, -1:A时间戳小于B时间粗 | 0：A等于B | 1：A大于B
	if order.Deadline.Compare(now) == -1 {
		_, err := dao.UpdateOrder(
			order.Id,
			dao.ORDER_PENDING,
			&dao.Orders{
				Status: "EXPIRED",
			},
		)
		if err != nil {
			logger.Error(err)
			logger.Infof("[First Queue] Order[%d]消息处理失败, 重新推回队列", order.Id)
			_ = delivery.Reject()
			return
		}

		logger.Infof("[First Queue] EXPIRED Order[%d]消息处理完毕", order.Id)
		_ = delivery.Ack()
		return
	}

	// 检测子钱包余额
	balance, err := checkTempWalletBalance(order)
	if err != nil {
		logger.Warnf("[First Queue] Order[%d]子钱包余额为0, 重新执行余额检测, err: %v", order.Id, err)
		_ = delivery.Reject()
		return
	}

	// First Queue结束TAG
	logger.Infof(
		"Order[%d]子钱包已收到%.8f %s, ",
		order.Id,
		balance,
		order.OriginalToken,
	)

	// 同步数据库并更新Transaction详情
	// TODO 待更新Transaction详情
	receivedHash, gcFloat, err := syncReceivedTx(
		hashOrder.Order.WeReceiveAddress,
		order.OriginalToken,
		order.OriginalTokenAmount,
		hashOrder.InitHeight,
	)
	if err != nil {
		logger.Warnf("[First Queue] Order[%d]检测收款详情失败, 重新推回队列, err: %v", order.Id, err)
		_ = delivery.Reject()
		return
	}

	receivedTxInfo := fmt.Sprintf(
		"%s*%d*%s*%s*%.18f",
		receivedHash,
		now.UnixMilli(),
		hashOrder.Order.OriginalToken,
		hashOrder.Order.OriginalTokenAmount,
		gcFloat,
	)

	// 更新数据库RECEIVED
	_, err = dao.UpdateOrder(
		hashOrder.Order.Id,
		dao.ORDER_PENDING,
		&dao.Orders{
			Status:         dao.ORDER_RECEIVED,
			ReceivedTxInfo: receivedTxInfo,
		})
	if err != nil {
		logger.Errorf("Order[%d] update failed, err: %v", hashOrder.Order.Id, err)
		_ = delivery.Reject()

		return
	}

	// 更新 LqLock
	incrRedisLqByOrder(hashOrder.Order, false)

	// 先推送到Second Queue，再进行Core队列分流，防止数据丢失
	dataByte, _ := json.Marshal(hashOrder)
	err = redis.SecondQueue.PublishBytes(dataByte)
	if err != nil {
		logger.Errorf("[First Queue] Order[%d] 推送第二队列失败, err: %v", order.Id, err)
		_ = delivery.Reject()
		return
	}

	// 队列消息处理完毕后, Ack() 处理
	_ = delivery.Ack()
}
