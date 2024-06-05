package queue

import (
	"collection-center/config"
	cnt "collection-center/contract/constant"
	"collection-center/internal/btc"
	"collection-center/internal/logger"
	"collection-center/internal/rpc"
	"collection-center/library/constant"
	"collection-center/library/redis"
	"collection-center/library/utils"
	"collection-center/service"
	"collection-center/service/db/dao"
	"encoding/json"
	"fmt"
	"github.com/adjust/rmq/v5"
	"github.com/ethereum/go-ethereum/core/types"
	orgRedis "github.com/redis/go-redis/v9"
	"math/big"
	"strconv"
	"time"
)

type ThirdConsumer struct {
	rmq.Consumer
}

func ThirdQueueConsumer() error {
	logger.Info("ThirdQueueConsumer StartConsuming 1")
	prefetchLimit := config.Config().Api.QueuePrefetchLimit
	if prefetchLimit == 0 {
		prefetchLimit = 100
	}
	err := redis.ThirdQueue.StartConsuming(int64(prefetchLimit), time.Second)
	if err != nil {
		logger.Error("ThirdQueueConsumer err:", err)
		return err
	}

	// 创建多个Consumer
	for i := 0; i < COSUMER_LIMIT; i++ {
		ret, err := redis.ThirdQueue.AddConsumer("ThirdQueueConsumer"+strconv.Itoa(i), new(ThirdConsumer))
		logger.Infof("第%d个consumer:%s添加成功[Second Queue]", i, ret)
		if err != nil {
			logger.Errorf("ThirdQueueConsumer err:%s", err)
			return err
		}
	}

	go func() {
		// push back the reject message time by time
		for {
			_, err := redis.ThirdQueue.ReturnRejected(RETURN_LIMIT)
			if err != nil {
				logger.Error(err)
				return
			}
			//logger.Warn("FirstQueue 恢复 ", rejected, " 条 reject 消息")
			time.Sleep(10 * time.Second)
		}
	}()

	logger.Info("ThirdQueueConsumer已启动")
	return nil
}

func (consumer *ThirdConsumer) Consume(delivery rmq.Delivery) {
	// 提取队列数据
	payload := delivery.Payload()
	payloadByte := []byte(payload)

	var hashOrder service.HashOrder
	err := json.Unmarshal(payloadByte, &hashOrder)
	if err != nil {
		logger.Error(err)
		logger.Warnf("[Third Queue] Order[%d]消息处理失败, 重新推回队列", hashOrder.Order.Id)
		_ = delivery.Reject()

		return
	}

	logger.Infof(
		"[Third Queue] order[%v To %v]:%+v\n",
		hashOrder.Order.OriginalToken,
		hashOrder.Order.TargetToken,
		hashOrder,
	)

	logger.Infof(
		"[Third Queue] Order[%d][%v To %v] is processing",
		hashOrder.Order.Id,
		hashOrder.Order.OriginalToken,
		hashOrder.Order.TargetToken,
	)

	if hashOrder.SendHash == "" {
		logger.Warnf("[Third Queue] Order[%d] SendHash is empty, 致命错误, 需要手动处理", hashOrder.Order.Id)
		_ = delivery.Ack()
		// 记录 error log
		err = dao.InsertErrorLogByOrder(hashOrder.Order.Id, hashOrder.Order.Status, hashOrder.Order.TargetToken, hashOrder.SendHash, fmt.Errorf("SendHash is empty"))
		if err != nil {
			logger.Error("[Third Queue] 插入失败日志失败: ", err)
		}
		return
	}

	var secondGas *big.Float

	if hashOrder.Order.TargetToken == "ETH" || hashOrder.Order.TargetToken == "USDT" {
		// 检查链上数据
		b, err := redis.GetChainData(cnt.ETH_HEIGHT)
		if err != nil {
			logger.Error(err)
			logger.Warnf("[Third Queue] Order[%d]消息处理失败, 重新推回队列, error: %v", hashOrder.Order.Id, err)
			_ = delivery.Reject()

			return
		}

		nowBlock, err := strconv.ParseUint(b, 10, 64)
		if err != nil {
			logger.Error("[Third Queue] GetChainData(cnt.ETH_HEIGHT) Invalid block Height, err:", err)
			_ = delivery.Reject()

			return
		}

		waitHeight := hashOrder.SendHeight + uint64(rpc.WaitBlock)

		logger.Infof("[Third Queue]Now ETH block Height %d from redis", nowBlock)

		if nowBlock < waitHeight {
			logger.Warnf("[Third Queue]Order ID[%d] Waiting block Height:%d, execute next round", hashOrder.Order.Id, waitHeight)
			_ = delivery.Reject()

			return
		}

		_, isPending, err := syncPendingEthReceipt(hashOrder.SendHash)
		if err != nil {
			logger.Errorf("[Third Queue] Order ID[%d] syncPendingEthReceipt error:%s, hash: %s", hashOrder.Order.Id, err, hashOrder.SendHash)
			_ = delivery.Reject()

			return
		}

		if isPending { // 在 memPool 中 --> 落块了之后, isPending -> false
			logger.Warnf("[Third Queue] Order ID[%d] [%v To %v] ETH-hash状态 Pending，推回Third Queue", hashOrder.Order.Id, hashOrder.Order.OriginalToken, hashOrder.Order.TargetToken)
			_ = delivery.Reject()
			return
		}

		rpt, err := syncEthReceipt(hashOrder.SendHash)
		if err != nil {
			// 校验高度更新次数
			// TODO 默认设置为2
			var MAX_UPDATE_COUNT uint64 = 2
			if hashOrder.UpdateHeightCount > MAX_UPDATE_COUNT {
				// resetOrder 处理了 Reject() 和 Ack()
				err = resetOrder(dao.ORDER_SENDING, dao.ORDER_RECEIVED, hashOrder, delivery)
				if err != nil {
					logger.Errorf("Reset order error:%v", err)
				}
				return
			}

			// 变更检测区块高度
			hashOrder.SendHeight = nowBlock
			hashOrder.UpdateHeightCount += 1

			err = pushTo3rdQueue(hashOrder)
			if err != nil {
				_ = delivery.Reject()
				return
			}

			_ = delivery.Ack()
			return
		}

		if rpt.Status != types.ReceiptStatusSuccessful {
			// resetOrder 处理了 Reject() 和 Ack()
			err = resetOrder(dao.ORDER_SENDING, dao.ORDER_RECEIVED, hashOrder, delivery)
			if err != nil {
				logger.Errorf("Reset order error:%v", err)
			}
			return
		}

		// 重新计算第二笔Transaction fee
		secondGas, err = SyncEthGasFee(hashOrder.SendHash)
		if err != nil {
			logger.Errorf("[Third Queue]Order[%d] Sync gas fee error:%v", hashOrder.Order.Id, err)
		} else {
			logger.Debugf("[Third Queue]Order[%d] ETH Sync gas fee:%v", hashOrder.Order.Id, secondGas.Text('f', -1))
		}

	} else if hashOrder.Order.TargetToken == "BTC" {
		// 检查链上数据
		b, err := redis.GetChainData(cnt.BTC_HEIGHT)
		if err != nil {
			logger.Error(err)
			logger.Warnf("[Third Queue]Order[%d]消息处理失败, 重新推回队列, err:%v", hashOrder.Order.Id, err)
			_ = delivery.Reject()

			return
		}

		nowBlock, err := strconv.ParseUint(b, 10, 64)
		if err != nil {
			logger.Error("[Third Queue] Invalid block Height, err:", err)
			_ = delivery.Reject()

			return
		}

		waitHeight := hashOrder.SendHeight + uint64(btc.WaitBlock)

		logger.Debugf("[Third Queue] Now BTC block Height %d from redis", nowBlock)

		if nowBlock < waitHeight {
			logger.Warnf("[Third Queue] Order ID[%d] Waiting BTC block Height:%d, execute next round", hashOrder.Order.Id, waitHeight)
			_ = delivery.Reject()

			return
		}

		// 平台BTC钱包确认需要使用第二个状态
		_, status2, gas, err := btc.GetTxStats(hashOrder.SendHash)
		if err != nil {
			logger.Errorf("BTC 发送用户资产失败:%v", err)
			_ = delivery.Reject()

			return
		}

		logger.Warnf("[Third Queue] Second hash[BTC] error:%v\n", status2)

		if !status2 {
			logger.Warnf("[Third Queue] Order[%v To %v] BTC-hash状态检测失败，推回Third Queue", hashOrder.Order.OriginalToken, hashOrder.Order.TargetToken)
			_ = delivery.Reject()

			return
		}

		// 重新计算第二笔gas
		secondGas = utils.BtcSatToB(gas)
	}

	closedTxInfo := fmt.Sprintf(
		"%s*%d*%s*%s*%s*%.18f",
		hashOrder.SendHash,
		time.Now().UnixMilli(),
		hashOrder.Order.TargetToken,
		hashOrder.Order.TargetTokenAmount,
		hashOrder.Order.TargetTokenReceived,
		secondGas,
	)
	updated := time.Now()
	// 更新数据库COMPLETE
	_, err = dao.UpdateOrder(hashOrder.Order.Id,
		dao.ORDER_SENDING,
		&dao.Orders{
			Status:              "COMPLETED",
			TargetTokenReceived: hashOrder.Order.TargetTokenReceived,
			ClosedTxInfo:        closedTxInfo,
			UpdatedAt:           updated,
		})
	if err != nil {
		logger.Errorf("[Third Queue] Order[%d] update failed, err: %v", hashOrder.Order.Id, err)
		_ = delivery.Reject()

		return
	}
	// 已完成订单信息, 写入 redis zset 中
	err = redis.Client().ZAdd(constant.OrderValueAll, orgRedis.Z{
		Score:  float64(updated.Unix()),
		Member: hashOrder.Order.OriginalTokenToU + "_" + strconv.FormatInt(hashOrder.Order.Id, 10),
	}).Err()
	if err != nil {
		logger.Error("[Third Queue] ZAdd error: ", err)
	}

	logger.Infof("[Third Queue] Order[%d] 已执行完成，退出当前第3队列", hashOrder.Order.Id)

	// 队列消息处理完毕后, Ack() 处理
	_ = delivery.Ack()
}
