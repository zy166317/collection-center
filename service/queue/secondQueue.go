package queue

import (
	"collection-center/config"
	cnt "collection-center/contract/constant"
	"collection-center/internal/btc"
	"collection-center/internal/logger"
	"collection-center/internal/rpc"
	"collection-center/library/redis"
	"collection-center/library/utils"
	"collection-center/service"
	"collection-center/service/db"
	"collection-center/service/db/dao"
	"context"
	"encoding/json"
	"github.com/adjust/rmq/v5"
	"github.com/ethereum/go-ethereum/common"
	"strconv"
	"time"
)

type SecondConsumer struct {
	rmq.Consumer
}

func SecondQueueConsumer() error {
	logger.Info("SecondQueueConsumer StartConsuming 1")
	prefetchLimit := config.Config().Api.QueuePrefetchLimit
	if prefetchLimit == 0 {
		prefetchLimit = 100
	}
	err := redis.SecondQueue.StartConsuming(int64(prefetchLimit), time.Second)
	if err != nil {
		logger.Error("SecondQueueConsumer err:", err)
		return err
	}

	// 创建多个Consumer
	for i := 0; i < COSUMER_LIMIT; i++ {
		ret, err := redis.SecondQueue.AddConsumer("SecondQueueConsumer"+strconv.Itoa(i), new(SecondConsumer))
		logger.Infof("第%d个consumer:%s添加成功[Second Queue]", i, ret)
		if err != nil {
			logger.Errorf("SecondQueueConsumer err:%s", err)
			return err
		}
	}

	go func() {
		// push back the reject message time by time
		for {
			_, err := redis.SecondQueue.ReturnRejected(RETURN_LIMIT)
			if err != nil {
				logger.Error(err)
				return
			}
			//logger.Warn("FirstQueue 恢复 ", rejected, " 条 reject 消息")
			time.Sleep(10 * time.Second)
		}
	}()

	logger.Info("SecondQueueConsumer已启动")
	return nil
}

// Second队列归集子钱包余额
func (consumer *SecondConsumer) Consume(delivery rmq.Delivery) {
	// 提取队列数据
	payload := delivery.Payload()
	payloadByte := []byte(payload)

	var hashOrder service.HashOrder
	err := json.Unmarshal(payloadByte, &hashOrder)
	if err != nil {
		logger.Error(err)
		logger.Warnf("[Second Queue]Order[%d]数据解析失败, 重新推回队列, err: %v", hashOrder.Order.Id, err)
		_ = delivery.Reject()

		return
	}

	// 推送到Core队列进行分流
	if !hashOrder.CoreToUser {
		// 推送到Core队列
		dataByteTmp, _ := json.Marshal(hashOrder)
		err = redis.CoreToUserQueue.PublishBytes(dataByteTmp)
		if err != nil {
			logger.Errorf("[Second Queue] Order[%d] 推送第Core队列失败, err: %v", hashOrder.Order.Id, err)
			_ = delivery.Reject()

			return
		}
		// 设置hashOrder数据为已执行核心钱包发送操作
		hashOrder.CoreToUser = true

		// 修改hashOrder后推回到Second队列
		dataByte, _ := json.Marshal(hashOrder)
		err = redis.SecondQueue.PublishBytes(dataByte)
		if err != nil {
			logger.Errorf("[Second Queue] Order[%d] 推送第二队列失败, err: %v", hashOrder.Order.Id, err)
			_ = delivery.Reject()
			return
		}

		// 更新队列数据，重新推回队列
		_ = delivery.Ack()
		return
	}

	logger.Infof(
		"[Second Queue]Order[%d][%v To %v] is collecting balance",
		hashOrder.Order.Id,
		hashOrder.Order.OriginalToken,
		hashOrder.Order.TargetToken,
	)

	// 进行订单子钱包余额归集
	order := hashOrder.Order
	if order.OriginalToken == "USDT" {
		if hashOrder.GasHash == "" {
			// TODO 检查余额,并确认是否大于最低gas标准
			// TODO gas最低标准计算

			hashOrder.GasHash, err = applyForGas(order.WeReceiveAddress)
			if err != nil {
				logger.Errorf("[Second Queue] USDT order apply for gas error:%v", err)
				_ = delivery.Reject()
				return
			}

			dataByte, _ := json.Marshal(hashOrder)
			err = redis.SecondQueue.PublishBytes(dataByte)
			if err != nil {
				logger.Errorf("[Second Queue] Order[%d] 推送第二队列失败, err: %v", hashOrder.Order.Id, err)
				_ = delivery.Reject()
				return
			}

			_ = delivery.Ack()
			return
		}

		ethRpc, err := rpc.NewEthRpc()
		if err != nil {
			logger.Errorf("[Second Queue] Order[%d] NewEthRpc error:%v", order.Id, err)
			_ = delivery.Reject()
			return
		}

		logger.Infof("[Second Queue] Order[%d] Gas hash:%s", order.Id, hashOrder.GasHash)

		tx := common.HexToHash(hashOrder.GasHash)
		// 确认gas hash 链上状态
		rep, err := ethRpc.SyncTxReceipt(context.Background(), &tx)
		if err != nil {
			logger.Errorf("[Second Queue] Order[%d] Tx receipt syncing status:%v", order.Id, err)
			_ = delivery.Reject()
			return
		}

		if rep.Status != 1 {
			// 上链成功，但显示失败
			// 恢复gas hash为空值，重新申请gas
			hashOrder.GasHash = ""

			dataByte, _ := json.Marshal(hashOrder)
			err = redis.SecondQueue.PublishBytes(dataByte)
			if err != nil {
				logger.Errorf("[Second Queue] Order[%d] 推送第二队列失败, err: %v", hashOrder.Order.Id, err)
				_ = delivery.Reject()
				return
			}

			_ = delivery.Ack()
			return
		}
	}

	// 判断归集等待时间周期 & gas成本
	if order.OriginalToken == "ETH" || order.OriginalToken == "USDT" {
		err := syncWaitBlockFirst(hashOrder, cnt.ETH_HEIGHT, rpc.WaitBlockCltEthMin, rpc.WaitBlockCltEthMax)
		if err != nil {
			logger.Error(err)
			_ = delivery.Reject()

			return
		}

		// 获取实时gas price
		ethRpc, err := rpc.NewEthRpc()
		if err != nil {
			logger.Error("NewEthRpc error:", err)
		}

		// 单位：Gwei
		gasPrice, err := ethRpc.SuggestGasPrice(context.Background())
		if err != nil {
			logger.Error(err)
			_ = delivery.Reject()

			return
		}

		logger.Warnf("Latest gas price:%v, max gas price:%v", gasPrice, rpc.EthMaxGasPrice)

		if gasPrice.Cmp(rpc.EthMaxGasPrice) == 1 {
			logger.Error("Gas price beyond default max limit")
			_ = delivery.Reject()

			return
		}
	} else if order.OriginalToken == "BTC" {
		err := syncWaitBlockFirst(hashOrder, cnt.BTC_HEIGHT, btc.WaitBlockCltBtcMin, btc.WaitBlockCltBtcMax)
		if err != nil {
			logger.Error(err)
			_ = delivery.Reject()

			return
		}
	}

	// 归集Core钱包
	if hashOrder.CollectedHash == "" {
		amount, _ := utils.StrToBigFloat(order.OriginalTokenAmount)
		hash, height, err := collectToCore(order, amount)
		if err != nil {
			// TODO 当前逻辑不需要用到
			//if order.OriginalToken == "USDT" && strings.Contains(err.Error(), "insufficient funds for gas") {
			//	logger.Warnf("[Second Queue] Order[%d]归集Core钱包失败, 重新申请Gas, err: %v", order.Id, err)
			//	// gas 不足的情况下, 清空 gas hash, 重新申请gas
			//	hashOrder.GasHash = ""
			//	err = repushByAck(hashOrder)
			//	if err != nil {
			//		logger.Errorf("[First Queue] Order[%d] 重新推送, err: %v", order.Id, err)
			//		_ = delivery.Reject()
			//		return
			//	}
			//	_ = delivery.Ack()
			//	return
			//}

			logger.Errorf("[Second Queue] Order[%d]归集Core钱包失败, 重新推回队列, err: %v", order.Id, err)

			_ = delivery.Reject()
			return
		}

		hashOrder.CollectedHeight = height
		hashOrder.CollectedHash = hash

		// 归集广播成功后，重新推回Second队列
		dataByte, _ := json.Marshal(hashOrder)
		err = redis.SecondQueue.PublishBytes(dataByte)
		if err != nil {
			logger.Errorf("[Second Queue] Order[%d] 推送第二队列失败, err: %v", hashOrder.Order.Id, err)
			_ = delivery.Reject()
			return
		}

		// 更新队列数据，重新推回队列
		_ = delivery.Ack()
		return
	}

	// 检查链上数据 - 是否等待足够的区块确认(归集钱包操作)
	if order.OriginalToken == "ETH" || order.OriginalToken == "USDT" {
		// 检查链上数据
		ethBlockHeightStr, err := redis.GetChainData(cnt.ETH_HEIGHT)
		if err != nil {
			logger.Error(err)
			logger.Warnf("[Second Queue]Order[%d]消息处理失败, 重新推回队列, err: %v", hashOrder.Order.Id, err)
			_ = delivery.Reject()

			return
		}

		nowBlock, err := strconv.ParseInt(ethBlockHeightStr, 10, 64)
		if err != nil {
			logger.Error("[Second Queue] Invalid block Height, err:", err)
			_ = delivery.Reject()

			return
		}

		waitHeight := hashOrder.CollectedHeight + uint64(rpc.WaitBlock)

		logger.Infof("[Second Queue]Now ETH block Height %d from redis", nowBlock)

		if uint64(nowBlock) < waitHeight {
			logger.Warnf("[Second Queue]Order ID[%d] Waiting block Height:%d, execute next round", hashOrder.Order.Id, waitHeight)
			_ = delivery.Reject()

			return
		}

		rpt, err := syncEthReceipt(hashOrder.CollectedHash)
		if err != nil {
			logger.Errorf("[Second Queue] Receipt error:%s, hash: %s", err, hashOrder.CollectedHash)
			_ = delivery.Reject()

			return
		}

		logger.Infof("[Second Queue] Transaction hash status:%v", rpt.Status)

		if rpt.Status != 1 {
			logger.Errorf("[Second Queue] Order[%v To %v] hash状态检测失败，推回Second Queue", order.OriginalToken, order.TargetToken)

			// 升级处理逻辑, 重新推回到队列
			_ = delivery.Reject()
			return
		}
	} else if order.OriginalToken == "BTC" {
		// 检查链上数据
		btcHeightStr, err := redis.GetChainData(cnt.BTC_HEIGHT)
		if err != nil {
			logger.Error(err)
			logger.Warnf("[Second Queue]Order[%d]消息处理失败, 重新推回队列, err: %v", hashOrder.Order.Id, err)
			_ = delivery.Reject()

			return
		}

		nowBlock, err := strconv.ParseInt(btcHeightStr, 10, 64)
		if err != nil {
			logger.Error("Invalid block Height")
			_ = delivery.Reject()

			return
		}

		waitHeight := hashOrder.CollectedHeight + uint64(btc.WaitBlock)

		logger.Infof("[Second Queue]Now BTC block Height %d from redis", nowBlock)

		if uint64(nowBlock) < waitHeight {
			logger.Warnf("[Second Queue]Order ID[%d] Waiting BTC block Height:%d, execute next round", hashOrder.Order.Id, waitHeight)
			_ = delivery.Reject()

			return
		}

		// 用户发出btc需要使用第一个bool值确认
		status1, _, _, err := btc.GetTxStats(hashOrder.CollectedHash)

		if err != nil {
			logger.Errorf("用户发送到子钱包：%v", err)
			_ = delivery.Reject()

			return
		}

		if !status1 {
			logger.Warnf("Order[%v To %v] hash状态检测失败，推回Second Queue", order.OriginalToken, order.TargetToken)
			_ = delivery.Reject()

			return
		}
	}

	logger.Infof("Order ID[%d]归集成功, Block-height:%d, TX-hash:%s",
		hashOrder.Order.Id,
		hashOrder.CollectedHeight,
		hashOrder.CollectedHash,
	)

	// 写入归集数据
	session := db.Client().NewSession()
	defer session.Close()

	err = dao.InsertCollectStep(
		session,
		hashOrder.Order.Id,
		hashOrder.CollectedHash,
	)
	if err != nil {
		logger.Errorf("Insert collect step table error:%v", err)

		_ = delivery.Reject()
		return
	}

	// 更新 LqLock
	incrRedisLqByOrder(hashOrder.Order, true)

	// 队列消息处理完毕后, Ack() 处理
	_ = delivery.Ack()
}
