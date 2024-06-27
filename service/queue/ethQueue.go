package queue

import (
	"collection-center/config"
	"collection-center/internal/logger"
	"collection-center/internal/rpc"
	"collection-center/library/redis"
	"collection-center/library/utils"
	"context"
	"encoding/json"
	"github.com/adjust/rmq/v5"
	"math"
	"math/big"
	"strconv"
	"time"
)

const COSUMER_LIMIT = 10
const RETURN_LIMIT = math.MaxInt16

type EthConsumer struct {
	rmq.Consumer
}

func EthQueueConsumer() error {
	logger.Info("ETH Queue Consumer Starting")
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

	logger.Info("EthQueueConsumer ETH队列消费者已启动")
	return nil
}

func (consumer *EthConsumer) Consume(delivery rmq.Delivery) {
	payload := delivery.Payload()
	payloadByte := []byte(payload)
	verifyOrder := &VerifyOrder{}
	err := json.Unmarshal(payloadByte, verifyOrder)
	if err != nil {
		logger.Error("EthQueueConsumer 消费失败:", err)
		_ = delivery.Reject()
		return
	}

	ethRpc, err := rpc.NewBaseEthRpc(false)
	if err != nil {
		logger.Error("NewBaseEthRpc:", err)
		_ = delivery.Reject()
		return
	}
	//定义交易金额，erc20 token addr
	var amount string
	var fromAddr string
	var toAddr string
	var tokenAddr string
	transaction, err := ethRpc.GetTransactionByTxSign(context.Background(), verifyOrder.Hash)
	if err != nil {
		logger.Error("GetTransferByTxSign:", err)
		_ = delivery.Reject()
		return
	}
	//定义erc20 decimals
	precision := utils.GetDecimalsInt(18)
	if transaction.Value == "0x0" {
		rc, err := ethRpc.SyncTransactionReceipt(context.Background(), transaction.Hash)
		if err != nil {
			logger.Error("SyncTransactionReceipt:", err)
			_ = delivery.Reject()
			return
		}
		//分析交易日志
		if len(rc.Logs) > 0 {
			newEthRpc, err := rpc.NewEthRpc()
			if err != nil {
				logger.Error("NewEthRpc:", err)
				_ = delivery.Reject()
				return

			}
			//获取代币精度
			precision, _, _, err = newEthRpc.GetTokenInfo(rc.Logs[len(rc.Logs)-1].Address)
			if err != nil {
				logger.Error("GetTokenDecimal error:", err)
				return
			}
			//根据代币合约获取收款地址
			toAddr = utils.Hex(rc.Logs[len(rc.Logs)-1].Topics[len(rc.Logs[0].Topics)-1])
			//TODO 校验交易收款地址
			if toAddr == "" {
				logger.Info("verify passed")
			} else {
				logger.Error("verify not passed", verifyOrder.Hash)
				return
			}
			//交易数量
			amount = rc.Logs[len(rc.Logs)-1].Data
			tokenAddr = rc.To
		}
	} else {
		//校验收款地址
		//if transaction.To != "" {
		//	logger.Info("校验通过")
		//} else {
		//	logger.Error("交易收款地址校验失败")
		//}
		amount = transaction.Value
	}
	//获取精度值
	s := precision.String()
	tokenPrecision := utils.GetDecimalsString(s)
	amountstr := utils.HexadecimalToString(amount)
	bigInt := new(big.Int)
	setString, _ := bigInt.SetString(amountstr, 10)
	//根据对应货币精度计算
	div := utils.DecimalsDiv(setString, tokenPrecision)
	//from to
	_ = fromAddr
	_ = toAddr
	//erc20 token address
	_ = tokenAddr
	logger.Info("test success", div)
	_ = delivery.Ack()

}

// eth转账||erc20标准Token校验
