package queue

import (
	"collection-center/config"
	"collection-center/internal/logger"
	"collection-center/library/redis"
	"collection-center/service"
	"collection-center/service/db"
	"collection-center/service/db/dao"
	"encoding/json"
	"github.com/adjust/rmq/v5"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

type CoreToUserConsumer struct {
	rmq.Consumer
}

func CoreToUserQueueConsumer() error {
	logger.Info("CoreToUserQueueConsumer StartConsuming 1")
	prefetchLimit := config.Config().Api.QueuePrefetchLimit
	if prefetchLimit == 0 {
		prefetchLimit = 100
	}
	err := redis.CoreToUserQueue.StartConsuming(int64(prefetchLimit), time.Second)
	if err != nil {
		logger.Error("CoreToUserQueueConsumer err:", err)
		return err
	}

	// 创建多个Consumer
	for i := 0; i < COSUMER_LIMIT; i++ {
		ret, err := redis.CoreToUserQueue.AddConsumer("CoreToUserQueueConsumer"+strconv.Itoa(i), new(CoreToUserConsumer))
		logger.Infof("第%d个consumer:%s添加成功[CoreToUser]", i, ret)
		if err != nil {
			logger.Errorf("CoreToUserQueueConsumer err:%s", err)
			return err
		}
	}

	go func() {
		// push back the reject message time by time
		for {
			_, err := redis.CoreToUserQueue.ReturnRejected(RETURN_LIMIT)
			if err != nil {
				logger.Error(err)
				return
			}
			//logger.Warn("FirstQueue 恢复 ", rejected, " 条 reject 消息")
			time.Sleep(10 * time.Second)
		}
	}()

	logger.Info("CoreToUserQueue 已启动")
	return nil
}

func (consumer *CoreToUserConsumer) Consume(delivery rmq.Delivery) {
	// 提取队列数据
	payload := delivery.Payload()
	payloadByte := []byte(payload)

	var hashOrder service.HashOrder
	err := json.Unmarshal(payloadByte, &hashOrder)
	if err != nil {
		logger.Error(err)
		logger.Infof("[Core Queue] Order[%d] 消息处理失败, 重新推回Core队列, err: %v", hashOrder.Order.Id, err)
		_ = delivery.Reject()

		return
	}

	logger.Infof(
		"[Core Queue]Order[%d][%v To %v] is processing",
		hashOrder.Order.Id,
		hashOrder.Order.OriginalToken,
		hashOrder.Order.TargetToken,
	)

	// 查询数据库订单状态
	order, err := dao.SelectOrderByID(hashOrder.Order.Id)
	if err != nil {
		logger.Error(err)
		logger.Infof("[Core Queue] Order[%d] GetOrderById 处理失败, 重新推回Core队列, err: %v", hashOrder.Order.Id, err)
		_ = delivery.Reject()

		return
	}

	// 已经存在 SendHash || 如果是 SENDING 状态, 说明已经处理过了, 直接 Ack() 处理
	if hashOrder.SendHash != "" || order.Status == dao.ORDER_SENDING {
		logger.Infof("[Core Queue] Order[%d] 已经处理过了, 直接 Ack() 处理, send hash: %s", hashOrder.Order.Id, hashOrder.SendHash)
		err = pushTo3rdQueue(hashOrder)
		if err != nil {
			_ = delivery.Reject()
			return
		}

		logger.Infof("[Core Queue] Order[%d][%v To %v] 已成功切换到第3队列", hashOrder.Order.Id, hashOrder.Order.OriginalToken, hashOrder.Order.TargetToken)
		_ = delivery.Ack()
		return
	}

	// 如果订单状态不是 RECEIVED, 则报错处理
	if order.Status != dao.ORDER_RECEIVED {
		logger.Errorf("[CoreToUser]Order[%d] status is not RECEIVED, status: %s, 需要手动处理, 退出队列: CoreToUser", hashOrder.Order.Id, order.Status)
		err := dao.InsertErrorLogByOrder(order.Id, order.Status, order.TargetToken, hashOrder.SendHash, errors.New("Order status is not RECEIVED in CoreToUserQueueConsumer"))
		if err != nil {
			logger.Error("[Core Queue] InsertErrorLogByOrder error:", err)
		}
		_ = delivery.Ack()
		return
	}

	// 更新数据库状态为 SENDING
	session := db.Client().NewSession()
	defer session.Close()
	session.Begin()
	err = dao.UpdateOrderStatusById(session, hashOrder.Order.Id, dao.ORDER_RECEIVED, dao.ORDER_SENDING)
	if err != nil {
		logger.Errorf("[Core Queue] Order[%d] update failed", hashOrder.Order.Id)
		session.Rollback()
		_ = delivery.Reject()
		return
	}

	// TODO 转账之前先查询是否已经转过了(无论成功与否, 防止重复转账)

	// core钱包转账给用户
	sendHash, received, height, err := sendOffSign(hashOrder)
	if err != nil {
		logger.Errorf("[Core Queue]Order[%d][%v To %v] is processing with SendOffSign error: %v", hashOrder.Order.Id, hashOrder.Order.OriginalToken, hashOrder.Order.TargetToken, err)
		session.Rollback()
		_ = delivery.Reject()
		return
	}

	if sendHash == "" {
		logger.Error("[Core Queue] Send hash error")
		session.Rollback()
		_ = delivery.Reject()
		return
	}

	logger.Infof("[Core Queue] Second hash:%v, Second height:%v, Received:%s", sendHash, height, received)

	hashOrder.SendHash = sendHash
	hashOrder.SendHeight = height
	hashOrder.Order.TargetTokenReceived = received

	err = session.Commit()
	if err != nil {
		logger.Error("[Core Queue] Commit error:", err)
		logger.Error("[Core Queue] OrderInfo: ", hashOrder)
		logger.Error("[Core Queue] 事务提交失败, 需要手动处理, order id:", hashOrder.Order.Id, "退出队列: CoreToUser")
		err := dao.InsertErrorLogByOrder(order.Id, order.Status, order.TargetToken, hashOrder.SendHash, errors.Wrap(err, "Commit error in CoreToUserQueueConsumer"))
		if err != nil {
			logger.Error("[Core Queue] InsertErrorLogByOrder error:", err)
		}
		_ = delivery.Ack()
		return
	}

	err = pushTo3rdQueue(hashOrder)
	if err != nil {
		logger.Errorf("[Core Queue] Order[%d] 推送第3队列失败, err: %v", hashOrder.Order.Id, err)
		_ = delivery.Reject()
		return
	}

	logger.Infof("[Core Queue] Order[%d][%v To %v] 已成功切换到第3队列", hashOrder.Order.Id, hashOrder.Order.OriginalToken, hashOrder.Order.TargetToken)

	// 队列消息处理完毕后, Ack() 处理
	_ = delivery.Ack()
}
