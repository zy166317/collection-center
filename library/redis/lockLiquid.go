package redis

import (
	"collection-center/internal/logger"
	"collection-center/library/constant"
	"collection-center/library/utils"
	"collection-center/service/db"
	"collection-center/service/db/dao"
	"math/big"
)

func InitLockLiquid() {
	logger.Info("=============InitLockLiquid start")
	initLockLiquid()
	logger.Info("=============InitLockLiquid end")
}

// initLockLiquid 初始化 流动性锁仓
func initLockLiquid() {
	liquidSlice := []string{ // 可以考虑加入配置文件
		constant.CoinEth,
		constant.CoinUsdt,
		constant.CoinBtc,
	}
	for _, coin := range liquidSlice {
		// 1. 清理 redis 锁仓
		Client().Del(constant.GetLqLockedKey(coin))
		// 2. 从 mysql 获取锁仓数据
		total, err := calculateSingleLq(coin)
		if err != nil {
			logger.Error("calculateSingleLq err:", err)
			return
		}
		// 3. 写入 redis
		_, err = InsertChainData(constant.GetLqLockedKey(coin), total.Text('f', -1))
		if err != nil {
			logger.Error("SetChainData err:", err)
			return
		}
	}
	return
}

// calculateSingleLq 从数据库获取锁仓数据
// 仅仅计算 dao.ORDER_RECEIVED 状态为锁仓
func calculateSingleLq(coin string) (*big.Float, error) {
	session := db.Client().NewSession()
	session = session.Where("(original_token = ? OR target_token = ?)", coin, coin)
	count, err := dao.CountOrderByStatus(session, 0, dao.ORDER_RECEIVED)
	if err != nil {
		logger.Error("CountOrderByStatus err:", err)
		return nil, err
	}
	if count == 0 {
		return big.NewFloat(0), nil
	}
	// 分页获取锁仓总量
	pageSize := 100
	page := 1
	total := big.NewFloat(0)
	session = db.Client().NewSession()
	session = session.
		Where("(original_token = ? OR target_token = ?)", coin, coin).
		And("status = ?", dao.ORDER_RECEIVED)
	for {
		if (page-1)*pageSize > int(count) {
			break
		}
		orders, err := dao.SelectOrders(session, page, pageSize)
		if err != nil {
			logger.Error("SelectOrders err:", err)
			return nil, err
		}
		for _, order := range orders {
			if order.OriginalToken == coin {
				amount, _ := utils.StrToBigFloat(order.OriginalTokenAmount)
				total = new(big.Float).Add(total, amount)
			} else if order.TargetToken == coin {
				amount, _ := utils.StrToBigFloat(order.TargetTokenAmount)
				total = new(big.Float).Add(total, amount)
			}
		}
		page++
	}
	return total, nil
}
