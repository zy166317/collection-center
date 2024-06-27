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
	rpc2 "github.com/gagliardetto/solana-go/rpc"
	"github.com/shopspring/decimal"
	"math/big"
	"strconv"
	"time"
)

type SolConsumer struct {
	rmq.Consumer
}

func SolQueueConsumer() error {
	logger.Info("Solana Queue Consumer Starting")
	prefetchLimit := config.Config().Api.QueuePrefetchLimit
	if prefetchLimit == 0 {
		prefetchLimit = 20
	}
	err := redis.SolQueue.StartConsuming(int64(prefetchLimit), 15*time.Second)
	if err != nil {
		logger.Error("SolQueueConsumer err:", err)
		return err
	}

	// 创建多个Consumer
	for i := 0; i < COSUMER_LIMIT; i++ {
		ret, err := redis.SolQueue.AddConsumer("SolQueueConsumer"+strconv.Itoa(i), new(SolConsumer))
		logger.Infof("第%d个consumer:%s添加成功", i, ret)
		if err != nil {
			logger.Errorf("SolQueueConsumer err:%s", err)
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
			_, err := redis.SolQueue.ReturnRejected(RETURN_LIMIT)
			if err != nil {
				logger.Error(err)
				return
			}
			//logger.Warn("FirstQueue 恢复 ", rejected, " 条 reject 消息")
			time.Sleep(20 * time.Second)
		}
	}()

	logger.Info("SolQueueConsumer Sol队列消费者已启动")
	return nil
}

func (consumer *SolConsumer) Consume(delivery rmq.Delivery) {
	payload := delivery.Payload()
	payloadByte := []byte(payload)
	verifyOrder := &VerifyOrder{}
	if err := json.Unmarshal(payloadByte, verifyOrder); err != nil {
		logger.Error("sol json.Unmarshal error:", err)
		_ = delivery.Reject()
		return
	}
	solRpc := rpc.NewSolRpc()
	//定义收款金额，收款地址，货币类型
	var toAmount string
	var toAddr string
	var tokenAddr string
	transaction, err := solRpc.GetSolTransaction(context.Background(), verifyOrder.Hash)
	if err != nil {
		logger.Error("solRpc.GetSolTransaction error:", err)
		_ = delivery.Reject()
		return
	}
	//
	//区分sol和spl
	if len(transaction.Meta.PreTokenBalances) > 0 {
		//spl 金额计算=发起人发送前金额-发起人发送后金额
		//获取收款地址收款前后金额
		preTokenBalance := rpc2.TokenBalance{}
		//获取spl转账信息，如果收款账户对应token为0，则不会出现转账钱的记录信息
		for _, v := range transaction.Meta.PreTokenBalances {
			preTokenBalance = v
			break
		}
		postTokenBalance := rpc2.TokenBalance{}
		for _, v := range transaction.Meta.PostTokenBalances {
			if v.AccountIndex == 1 {
				//收款地址
				toAddr = v.Owner.String()
			}
			if v.AccountIndex == preTokenBalance.AccountIndex {
				postTokenBalance = v
				break
			}
		}
		//无效spl转账交易
		if preTokenBalance.AccountIndex == 0 || postTokenBalance.AccountIndex == 0 {
			logger.Error("invalid spl transaction")
			return
		}
		pre, err := decimal.NewFromString(preTokenBalance.UiTokenAmount.UiAmountString)
		if err != nil {
			logger.Error("sol decimal.NewFromString error:", err)
			_ = delivery.Reject()
			return
		}
		post, err := decimal.NewFromString(postTokenBalance.UiTokenAmount.UiAmountString)
		if err != nil {
			logger.Error("sol decimal.NewFromString error:", err)
			_ = delivery.Reject()
			return
		}
		//获取token地址
		tokenAddr = postTokenBalance.Mint.String()
		//计算转账金额,转账金额取绝对值
		toAmount = pre.Sub(post).Abs().String()
	} else {
		//sol
		txdata, err := transaction.Transaction.GetTransaction()
		if err != nil {
			logger.Error("sol GetTransaction error:", err)
			_ = delivery.Reject()
			return
		}
		toAddr = txdata.Message.AccountKeys[1].String()
		//获取精度
		value := transaction.Meta.PostBalances[1] - transaction.Meta.PreBalances[1]
		toAmount = utils.DecimalsDiv(new(big.Int).SetUint64(value), new(big.Int).SetUint64(1000000000)).String()
	}
	//获取到账金额
	_ = toAmount
	logger.Info("to amount ", toAmount)
	//TODO 校验收款地址
	_ = toAddr
	//spl token地址
	_ = tokenAddr
	_ = delivery.Ack()
}
