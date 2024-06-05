package queue

import (
	cnt "collection-center/contract/constant"
	"collection-center/internal/btc"
	"collection-center/internal/logger"
	"collection-center/internal/rpc"
	"collection-center/library/constant"
	redis2 "collection-center/library/redis"
	"collection-center/library/utils"
	"collection-center/library/wallet"
	"collection-center/service"
	"collection-center/service/db"
	"collection-center/service/db/dao"
	"collection-center/service/price"
	"context"
	"encoding/json"
	"fmt"
	"github.com/adjust/rmq/v5"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"golang.org/x/xerrors"
	"math"
	"math/big"
	"strconv"
)

const COSUMER_LIMIT = 10
const RETURN_LIMIT = math.MaxInt16

func SyncEthGasFee(hashStr string) (*big.Float, error) {
	ethRpc, err := rpc.NewEthRpc()
	if err != nil {
		return nil, err
	}

	hash := common.HexToHash(hashStr)
	rep, err := ethRpc.SyncTxReceipt(context.Background(), &hash)
	if err != nil {
		return nil, err
	}

	decimals18, _ := utils.StrToBigFloat(cnt.DECIMALS_WEI)

	return new(big.Float).Quo(
		new(big.Float).SetInt(
			new(big.Int).Mul(big.NewInt(int64(rep.GasUsed)), rep.EffectiveGasPrice),
		),
		decimals18,
	), nil
}

func calculateLeftAmount(hashOrder service.HashOrder, sendType string) (*big.Float, *big.Float, error) {
	var lAmount *big.Float

	ethRpc, err := rpc.NewEthRpc()
	if err != nil {
		return nil, nil, err
	}

	tta, _ := utils.StrToBigFloat(hashOrder.Order.TargetTokenAmount)

	bFloatPrice, _, err := price.MultiTypeChainPrice()
	if err != nil {
		return nil, nil, err
	}

	if hashOrder.GasCost == nil {
		hashOrder.GasCost = big.NewFloat(0)
	}

	switch sendType {
	case "ETH":
		// 第一笔Gas
		gas1 := new(big.Float).Mul(hashOrder.GasCost, bFloatPrice.UsdtPerEth)

		// 第二笔Gas
		gasCostOrigin, err := ethRpc.GasCost(context.Background(), "ETH")
		if err != nil {
			return nil, nil, err
		}
		gas2, err := service.DecNumToReadNum(gasCostOrigin, 18)
		if err != nil {
			return nil, nil, err
		}

		if gas1.Cmp(big.NewFloat(0)) == 0 {
			gas1 = gas2
		}

		gasCost := new(big.Float).Add(gas1, gas2)
		lAmount = new(big.Float).Sub(tta, gasCost)

		return lAmount, gas2, nil
	case "USDT":
		// 第一笔gas
		gas1 := hashOrder.GasCost
		// 第二笔gas
		gasCostOrigin, err := ethRpc.GasCost(context.Background(), "ERC20")
		if err != nil {
			return nil, nil, err
		}
		gas2, err := service.DecNumToReadNum(gasCostOrigin, 18)
		if err != nil {
			return nil, nil, err
		}

		if gas1.Cmp(big.NewFloat(0)) == 0 {
			gas1 = gas2
		}

		gasCost := new(big.Float).Add(gas1, gas2)
		lAmount = new(big.Float).Sub(tta, gasCost)

		return lAmount, gas2, nil

	case "BTC":
		// 第一笔gas成本 USDT=>BTC
		gas1 := new(big.Float).Mul(hashOrder.GasCost, bFloatPrice.UsdtPerBtc)
		if gas1.Cmp(big.NewFloat(0)) == 0 {
			gasCostOrigin, err := ethRpc.GasCost(context.Background(), "ERC20")
			if err != nil {
				return nil, nil, err
			}
			intGas, err := service.EthDecToUSDTDec(gasCostOrigin.String())
			if err != nil {
				return nil, nil, err
			}
			// ETH => USDT
			gasUsd, err := service.DecNumToReadNum(intGas, 6)
			if err != nil {
				return nil, nil, err
			}

			gas1 = new(big.Float).Mul(gasUsd, bFloatPrice.UsdtPerBtc)
		}

		//// 第二笔gas成本
		gas2, _ := utils.StrToBigFloat(cnt.DEFAULT_BTC_GAS)

		gasCost := new(big.Float).Add(gas1, gas2)
		lAmount = new(big.Float).Sub(tta, gasCost)

		return lAmount, gas2, nil
	}

	return nil, nil, nil
}

// SendOffSign core钱包转账给用户
func sendOffSign(hashOrder service.HashOrder) (string, string, uint64, error) {
	ethRpc, err := rpc.NewEthRpc()
	if err != nil {
		return "", "", 0, err
	}

	// 重新计算FLOAT模式下最终兑换数量
	if hashOrder.Order.Mode == "FLOAT" {
		out, _, err := service.CalculateOut(
			hashOrder.Order.Mode,
			hashOrder.Order.OriginalToken,
			hashOrder.Order.OriginalTokenAmount,
			hashOrder.Order.TargetToken,
		)
		if err != nil {
			return "", "", 0, err
		}

		hashOrder.Order.TargetTokenAmount = out
	}

	var sendHash string
	var blockHeight uint64
	var receivedAmount string
	switch hashOrder.Order.TargetToken {
	case "ETH":
		lAmount, _, err := calculateLeftAmount(hashOrder, hashOrder.Order.TargetToken)
		if err != nil {
			return "", "", 0, err
		}

		sAmount, err := service.ReadNumToDecNum(lAmount, 18)
		if err != nil {
			return "", "", 0, err
		}

		receiver := common.HexToAddress(hashOrder.Order.UserReceiveAddress)

		tx, height, err := ethRpc.SendEthOffSign(sAmount, receiver)
		if err != nil {
			return "", "", 0, err
		}

		blockHeight = height
		sendHash = tx.String()
		receivedAmount = lAmount.Text('f', -1)

		break
	case "USDT":
		lAmount, _, err := calculateLeftAmount(hashOrder, hashOrder.Order.TargetToken)
		if err != nil {
			return "", "", 0, err
		}

		sAmount, err := service.ReadNumToDecNum(lAmount, 6)
		if err != nil {
			return "", "", 0, err
		}

		receiver := common.HexToAddress(hashOrder.Order.UserReceiveAddress)

		tx, height, err := ethRpc.SendERC20OffSign(sAmount, receiver, rpc.EvmAddrs.UsdtErc20)
		if err != nil {
			logger.Error("Send target token[USDT] error:", err, ", order ID:", hashOrder.Order.Id, ", receiver:", receiver.Hex(), ", amount:", sAmount.String())
			return "", "", 0, err
		}

		logger.Warnf("SendEthOffSign[target-USDT]:%v", tx.Hex())

		blockHeight = height
		sendHash = tx.String()
		receivedAmount = lAmount.Text('f', -1)

		break
	case "BTC":
		lAmount, _, err := calculateLeftAmount(hashOrder, hashOrder.Order.TargetToken)
		if err != nil {
			return "", "", 0, err
		}

		height, err := btc.Client.GetBlockCount()
		if err != nil {
			return "", "", 0, err
		}

		tx, err := btc.SendBTC(btc.BtcCoreWallet, hashOrder.Order.UserReceiveAddress, lAmount.String())
		if err != nil {
			logger.Errorf("Send target token[BTC] error:%v", err)
			return "", "", 0, err
		}

		logger.Warnf("SendEthOffSign[target-BTC]:%v", tx)

		blockHeight = uint64(height)
		sendHash = tx
		receivedAmount = lAmount.Text('f', -1)

		break
	}

	logger.Warnf("SendEthOffSign[target-%v]:%v", hashOrder.Order.TargetToken, sendHash)

	return sendHash, receivedAmount, blockHeight, nil
}

// syncEthReceipt 只同步已经上链的交易
func syncEthReceipt(hash string) (*types.Receipt, error) {
	tx := common.HexToHash(hash)
	// 归集到Core钱包
	ethRpc, err := rpc.NewEthRpc()
	if err != nil {
		return nil, err
	}

	rpt, err := ethRpc.SyncTxReceipt(context.Background(), &tx)
	if err != nil {
		return nil, err
	}

	return rpt, nil
}

// syncPendingEthReceipt 算上交易池内, 还未上链的交易
func syncPendingEthReceipt(hash string) (*types.Transaction, bool, error) {
	tx := common.HexToHash(hash)
	// 归集到Core钱包
	ethRpc, err := rpc.NewEthRpc()
	if err != nil {
		return nil, false, err
	}

	return ethRpc.SyncPendingTxReceipt(context.Background(), &tx)
}

func checkTempWalletBalance(order *dao.Orders) (*big.Float, error) {
	// 检验子钱包是否已到账
	var balance *big.Float

	receiveWallet := order.WeReceiveAddress
	originAmount, _ := utils.StrToBigFloat(order.OriginalTokenAmount)
	decimals18, _ := utils.StrToBigFloat(cnt.DECIMALS_WEI)
	decimalsUSDT, _ := utils.StrToBigFloat(cnt.DECIMALS_USDT)

	switch order.OriginalToken {
	case "ETH":
		ethRpc, err := rpc.NewEthRpc()
		if err != nil {
			logger.Error("checkTempWalletBalance: ", err)
			return nil, err
		}

		b, err := ethRpc.BalanceOfETH(receiveWallet)
		if err != nil {
			logger.Error("checkTempWalletBalance: ", err)
			return nil, err
		}
		balance, _ = utils.StrToBigFloat(b)
		balance = new(big.Float).Quo(balance, decimals18)
		break
	case "USDT":
		ethRpc, err := rpc.NewEthRpc()
		if err != nil {
			logger.Error("checkTempWalletBalance: ", err)
			return nil, err
		}

		b, _ := ethRpc.BalanceOfERC20(receiveWallet, rpc.EvmAddrs.UsdtErc20)
		if err != nil {
			logger.Error("checkTempWalletBalance: ", err)
			return nil, err
		}
		balance, _ = utils.StrToBigFloat(b)
		balance = new(big.Float).Quo(balance, decimalsUSDT)
		break
	case "BTC":
		a, err := btc.GetBalance(receiveWallet)
		if err != nil {
			logger.Error("checkTempWalletBalance: ", err)
			return nil, err
		}
		balance, _ = utils.StrToBigFloat(a)
		break
	default:
		return nil, xerrors.New("Invalid original token")
	}

	// 收到数量与订单数量做对比
	if balance.Cmp(originAmount) == -1 {
		return nil, xerrors.New(fmt.Sprintf("%s钱包还没收到转账", order.OriginalToken))
	}

	return balance, nil
}

func collectToCore(order *dao.Orders, balance *big.Float) (string, uint64, error) {
	var blockNum uint64
	var hashStr string

	// 提取子钱包私钥
	wlt, err := dao.SelectOrderByAddr(order.WeReceiveAddress)
	if err != nil {
		return "", 0, err
	}

	switch order.OriginalToken {
	case "ETH":
		// 归集到Core钱包
		ethRpc, err := rpc.NewEthRpc()
		if err != nil {
			return "", 0, err
		}

		num, err := ethRpc.Client.BlockNumber(context.Background())
		if err != nil {
			return "", 0, err
		}
		blockNum = num

		pvk, err := wallet.GenPvkObj(wlt.EncryptedKey)
		if err != nil {
			return "", 0, err
		}

		// 转为big.int
		decimals18, _ := utils.StrToBigFloat(cnt.DECIMALS_WEI)
		bln := new(big.Int)
		new(big.Float).Mul(balance, decimals18).Int(bln)

		// 扣除gas成本
		gasCost, err := ethRpc.GasCost(context.Background(), "ETH")
		if err != nil {
			return "", 0, err
		}
		leftAmount := new(big.Int).Sub(bln, gasCost)

		addr := common.HexToAddress(rpc.EthCoreWalletAddr)

		hash, err := ethRpc.SendTx(rpc.SendingInfo{
			PvKey:    pvk,
			Amount:   leftAmount,
			Receiver: addr,
		})
		if err != nil {
			return "", 0, err
		}

		hashStr = hash.String()

		break
	case "USDT":
		// 归集到Core钱包
		ethRpc, err := rpc.NewEthRpc()
		if err != nil {
			return "", 0, err
		}

		// 获取链上高度
		num, err := ethRpc.Client.BlockNumber(context.Background())
		if err != nil {
			return "", 0, err
		}
		blockNum = num

		// 计算gas
		decimals, _ := utils.StrToBigFloat(cnt.DECIMALS_USDT)
		sumBalance, _ := new(big.Float).Mul(balance, decimals).Int(new(big.Int))
		coreWallet := common.HexToAddress(rpc.EthCoreWalletAddr)

		pvk, err := wallet.GenPvkObj(wlt.EncryptedKey)
		if err != nil {
			return "", 0, err
		}
		fromAddr := wallet.GenWalletByKey(pvk)

		nonce, err := ethRpc.PendingNonce(fromAddr)

		hash, err := ethRpc.SendERC20(rpc.SendingInfo{
			PvKey:     pvk,
			Amount:    sumBalance,
			Receiver:  coreWallet,
			TokenAddr: rpc.EvmAddrs.UsdtErc20,
		}, nonce)
		if err != nil {
			return "", 0, err
		}
		hashStr = hash.String()

		break
	case "BTC":
		// 计算发送gas成本
		blockCount, err := btc.Client.GetBlockCount()
		if err != nil {
			return "", 0, err
		}
		blockNum = uint64(blockCount)

		// 扣除成本后的amount
		fee, _ := utils.StrToBigFloat(cnt.DEFAULT_BTC_GAS)
		originTokenAmount, _ := utils.StrToBigFloat(order.OriginalTokenAmount)
		leftAmount := new(big.Float).Sub(originTokenAmount, fee)

		tx, err := btc.SendBTC(order.WeReceiveAddress, btc.BtcCoreWallet, leftAmount.String())
		if err != nil {
			return "", 0, err
		}

		hashStr = tx

		break
	default:
		return "", 0, err
	}

	return hashStr, blockNum, nil
}

func incrRedisLqByOrder(order *dao.Orders, negative bool) {
	// 根据 order 情况 更新 流动性锁定信息
	logger.Debugf("incrRedisLqByOrder: negative %v, orderId %d, TargetTokenAmount: %v, targetToken: %v , OriginalTokenAmount: %v, originToken: %v",
		negative, order.Id, order.TargetTokenAmount, order.TargetToken, order.OriginalTokenAmount, order.OriginalToken)
	tta, _ := utils.StrToBigFloat(order.TargetTokenAmount)
	ttaF := utils.MustToFloat(tta)

	ota, _ := utils.StrToBigFloat(order.OriginalTokenAmount)
	otaF := utils.MustToFloat(ota)
	if negative {
		otaF = -1 * otaF
		ttaF = -1 * ttaF
		// 获取锁仓数据, 防止出现负数
		otaLocked, err := redis2.GetChainData(constant.GetLqLockedKey(order.OriginalToken))
		if err != nil {
			logger.Error("incrRedisLqByOrder: ", err)
			return
		}
		ttaLocked, err := redis2.GetChainData(constant.GetLqLockedKey(order.TargetToken))
		if err != nil {
			logger.Error("incrRedisLqByOrder: ", err)
			return
		}
		otaLockedF, _ := utils.StrToBigFloat(otaLocked)
		ttaLockedF, _ := utils.StrToBigFloat(ttaLocked)
		if otaLockedF.Cmp(ota) == -1 || ttaLockedF.Cmp(tta) == -1 {
			logger.Warnf("incrRedisLqByOrder: force to init LQ data, orderId[%d], LQ data: otaLocked %v, ttaLocked %v, negative to lose: ota: %v, tta: %v",
				order.Id, otaLocked, ttaLocked, order.OriginalTokenAmount, order.TargetTokenAmount)
			// 负数, 重新初始化数据
			redis2.InitLockLiquid()
			return
		}
	}
	redis2.Client().IncrByFloat(constant.GetLqLockedKey(order.TargetToken), ttaF)
	redis2.Client().IncrByFloat(constant.GetLqLockedKey(order.OriginalToken), otaF)
}

func applyForGas(receiveAddr string) (string, error) {
	ethRpc, err := rpc.NewEthRpc()
	if err != nil {
		return "", err
	}

	// 计算gas
	gasPrice, err := ethRpc.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}

	// 发 1.5 倍的 Gas, 用于归集, 防止 Gas 不足
	gas := new(big.Int).Mul(gasPrice, big.NewInt(int64(cnt.GASLIMIT_ERC20)))
	to := common.HexToAddress(receiveAddr)

	// 确保子钱包收到 Gas 用于归集
	tx, _, err := ethRpc.SendEthOffSign(gas, to)
	if err != nil {
		return "", err
	}

	return tx.String(), nil
}

func syncWaitBlockFirst(hashOrder service.HashOrder, chainType string, minWait int, maxWait int) error {
	ethBlockHeightStr, err := redis2.GetChainData(chainType)
	if err != nil {
		return err
	}

	nowBlock, err := strconv.ParseInt(ethBlockHeightStr, 10, 64)
	if err != nil {
		return err
	}

	randomHeight := utils.RangeRandom(minWait, maxWait)

	waitHeight := hashOrder.InitHeight + uint64(randomHeight)

	logger.Infof("[Second Queue]Now %v block Height %d from redis", hashOrder.Order.OriginalToken, nowBlock)

	if uint64(nowBlock) < waitHeight {
		return xerrors.New(fmt.Sprintf("[Second Queue]Order ID[%d] Waiting block Height:%d, execute next round", hashOrder.Order.Id, waitHeight))
	}

	return nil
}

func syncReceivedTx(addr string, coinType string, inAmount string, height uint64) (string, *big.Float, error) {
	var txHash string
	var gasCost *big.Float

	if coinType == "ETH" || coinType == "USDT" {
		client, err := rpc.NewEthRpc()
		if err != nil {
			return "", nil, err
		}

		txHash, gasCost, err = client.GetAddrTransfers(addr, int64(height), coinType, inAmount)
		if err != nil {
			return "", nil, err
		}
	} else if coinType == "BTC" {
		txs, err := btc.GetTxsByAddr(addr)
		if err != nil {
			return "", nil, err
		}

		if len(txs.TxRefs) == 0 {
			return "", nil, xerrors.New("Empty TXs array")
		}

		// TODO 默认校验第一条(待升级)
		tx := txs.TxRefs[0]
		// 获取该hash的金额
		val := tx.Value
		txHeight := tx.BlockHeight
		txHash = tx.TxHash

		// 高度校对 (订单起始height vs Tx height)
		if int64(height) > txHeight {
			return "", nil, xerrors.New("Oldest tx block height")
		}

		// 金额校对
		iaFloat, _ := utils.StrToBigFloat(inAmount)
		in, _ := service.ReadNumToDecNum(iaFloat, 8)
		if big.NewInt(val).Cmp(in) == -1 {
			return "", nil, xerrors.New("Not enough received value")
		}

		// 获取gas fee
		_, _, gas, _ := btc.GetTxStats(txHash)
		gasCost, err = service.DecNumToReadNum(big.NewInt(gas), 8)
		if err != nil {
			return "", nil, err
		}
	}

	return txHash, gasCost, nil
}

func pushTo3rdQueue(hashOrder service.HashOrder) error {
	dataByte, _ := json.Marshal(hashOrder)
	err := redis2.ThirdQueue.PublishBytes(dataByte)
	if err != nil {
		return err
	}
	return nil
}

// resetOrder 处理了 Reject() 和 Ack()
// 回滚订单
func resetOrder(curStatus string, updateStatus string, data service.HashOrder, delivery rmq.Delivery) error {
	// 判断需要升级的状态
	data.SendHash = ""
	data.SendHeight = 0
	data.Order.TargetTokenReceived = "0"
	data.Order.Status = dao.ORDER_RECEIVED

	// 变更数据库
	session := db.Client().NewSession()
	defer session.Close()
	session.Begin()
	err := dao.UpdateOrderStatusById(
		session,
		data.Order.Id,
		curStatus,
		updateStatus,
	)
	if err != nil {
		logger.Errorf("[Third Queue] Order[%d] reset update failed", data.Order.Id)

		session.Rollback()
		_ = delivery.Reject()
		return err
	}

	err = session.Commit()
	if err != nil {
		logger.Errorf("[Third Queue] Order[%d] reset update failed", data.Order.Id)

		session.Rollback()
		_ = delivery.Reject()
		return err
	}

	// 推回2.5队列
	dataByte, _ := json.Marshal(data)
	err = redis2.CoreToUserQueue.PublishBytes(dataByte)
	if err != nil {
		_ = delivery.Reject()
		return err
	}
	_ = dao.InsertErrorLogByOrder(data.Order.Id, curStatus, data.Order.TargetToken, data.SendHash, fmt.Errorf("reset order in third queue to 2.5 queue"))
	logger.Warnf("[Third Queue]Order[%d]交易状态查询失败, 重新推回2.5队列", data.Order.Id)

	_ = delivery.Ack()

	return nil
}
