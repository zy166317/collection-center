package service

import (
	cnt "collection-center/contract/constant"
	"collection-center/internal/btc"
	"collection-center/internal/ecode"
	"collection-center/internal/email"
	"collection-center/internal/logger"
	"collection-center/internal/rpc"
	"collection-center/library/constant"
	"collection-center/library/redis"
	"collection-center/library/request"
	"collection-center/library/utils"
	"collection-center/library/wallet"
	"collection-center/service/db"
	"collection-center/service/db/dao"
	"collection-center/service/price"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	orgRedis "github.com/redis/go-redis/v9"
	uuid "github.com/satori/go.uuid"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/xerrors"
)

type HashOrder struct {
	Order             *dao.Orders
	CoreToUser        bool   // 是否已执行转账给用户
	InitHeight        uint64 // 订单创建起始高度，用于检验余额 | 余额归集
	CollectedHeight   uint64
	CollectedHash     string
	GasHash           string
	GasCost           *big.Float
	SendHash          string
	SendHeight        uint64
	UpdateHeightCount uint64
	//MaxRetryCount int // 最大重试次数 5 次
}

type HistoryOrder struct {
	Id                  int64
	OriginalToken       string
	OriginalTokenAmount string
	TargetToken         string
	TargetTokenAmount   string
	CreatedAt           time.Time
	CompletedAt         string
}

func (o *HashOrder) MarshalToMsg() ([]byte, error) {
	bs, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func GenerateOrder(req *request.OrderReq) (*dao.Orders, error) {
	client, err := rpc.NewEthRpc()
	if err != nil {
		return nil, err
	}

	// 判断锁仓量
	var leftAmount *big.Float
	tta, _ := utils.StrToBigFloat(req.Targettokenamount)
	switch req.Targettoken {
	case "ETH":
		ETHLOCKED, err := redis.GetChainData(constant.GetLqLockedKey(constant.CoinEth))
		if err != nil && !errors.Is(err, orgRedis.Nil) {
			logger.Error("redis GetChainData CoinEth err:", err)
			return nil, err
		}
		ethLocked, _ := utils.StrToBigFloat(ETHLOCKED)

		// 获取钱包eth余额
		ethBalance, err := client.BalanceOfETH(rpc.EthCoreWalletAddr)
		if err != nil {
			return nil, err
		}
		decimals18, _ := utils.StrToBigFloat(cnt.DECIMALS_WEI)
		eBalaceFloat, _ := utils.StrToBigFloat(ethBalance)

		eBalaceFloat = new(big.Float).Quo(eBalaceFloat, decimals18)
		leftAmount = new(big.Float).Sub(eBalaceFloat, ethLocked)
		if leftAmount.Cmp(tta) != 1 {
			return nil, xerrors.New(fmt.Sprintf("%v输出数量不足，重新设置", req.Targettoken))
		}
		break
	case "USDT":
		USDTLOCKED, err := redis.GetChainData(constant.GetLqLockedKey(constant.CoinUsdt))
		if err != nil && !errors.Is(err, orgRedis.Nil) {
			logger.Error("redis GetChainData CoinUsdt err:", err)

			return nil, err
		}

		usdtLocked, _ := utils.StrToBigFloat(USDTLOCKED)

		// 获取钱包eth余额
		UsdtBalance, err := client.BalanceOfERC20(rpc.EthCoreWalletAddr, rpc.EvmAddrs.UsdtErc20)
		if err != nil {
			return nil, err
		}

		decimalsU, _ := utils.StrToBigFloat(cnt.DECIMALS_USDT)
		uBalanceFloat, _ := utils.StrToBigFloat(UsdtBalance)
		uBalanceFloat = new(big.Float).Quo(uBalanceFloat, decimalsU)
		leftAmount = new(big.Float).Sub(uBalanceFloat, usdtLocked)
		if leftAmount.Cmp(tta) != 1 {
			return nil, xerrors.New(fmt.Sprintf("%v输出数量不足，重新设置", req.Targettoken))
		}
		break
	case "BTC":
		BTCLOCKED, err := redis.GetChainData(constant.GetLqLockedKey(constant.CoinBtc))
		if err != nil && !errors.Is(err, orgRedis.Nil) {
			logger.Error("redis GetChainData CoinBtc err:", err)

			return nil, err
		}

		btcLocked, _ := utils.StrToBigFloat(BTCLOCKED)

		btcBalance, err := btc.GetBalance(btc.BtcCoreWallet)
		if err != nil {
			return nil, err
		}
		bBalance, _ := utils.StrToBigFloat(btcBalance)
		leftAmount = new(big.Float).Sub(bBalance, btcLocked)
		if leftAmount.Cmp(tta) != 1 {
			return nil, xerrors.New(fmt.Sprintf("%v输出数量不足，重新设置", req.Targettoken))
		}
		break
	}

	targetTokenAmount := req.Targettokenamount
	aAmount, err := utils.StrToBigFloat(targetTokenAmount)
	if err != nil {
		return nil, err
	}

	if leftAmount.Cmp(aAmount) != 1 {
		return nil, xerrors.New("核心钱包余额不足")
	}

	outAmount, firstGasCostDef, err := CalculateOut(req.Mode, req.Originaltoken, req.Originaltokenamount, req.Targettoken)
	if err != nil {
		return nil, err
	}
	bAmount, err := utils.StrToBigFloat(outAmount)
	if err != nil {
		return nil, err
	}

	spread := new(big.Float).Sub(aAmount, bAmount)
	spread = new(big.Float).Abs(spread)

	gapRate := new(big.Float).Quo(spread, aAmount)

	defaultGapRate, err := utils.StrToBigFloat(cnt.AMOUNT_GAP)
	if err != nil {
		return nil, err
	}

	// val等于1说明传进来的amount与计算值差距太大
	if gapRate.Cmp(defaultGapRate) == 1 {
		return nil, xerrors.New("Invalid target token amount")
	}

	var walletAddr string
	var encryedKey string
	var nowBlock uint64
	// 生成钱包

	// 重写生成钱包
	// 创建子钱包记录
	if req.Originaltoken == "ETH" || req.Originaltoken == "USDT" {
		pvk, _, addr, err := wallet.GenKeyEthWallet()
		if err != nil {
			return nil, err
		}
		key, err := wallet.GenPrivateKey(pvk)
		if err != nil {
			return nil, err
		}

		walletAddr = addr.String()
		encryedKey = key

		// 获取ETH区块高度
		nowBlock, err = redis.GetHeightFormRedis(cnt.ETH_HEIGHT)
		if err != nil {
			return nil, err
		}
	} else if req.Originaltoken == "BTC" {
		btcObj := btc.GenKeyBTCWallet()
		walletAddr = btcObj.Address
		encryedKey = btcObj.WIF

		nowBlock, err = redis.GetHeightFormRedis(cnt.BTC_HEIGHT)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, xerrors.New("Invalid target token")
	}

	// 写入wallet表
	encryedKeyStr, err := utils.Encrypt(encryedKey)
	if err != nil {
		return nil, err
	}
	_, err = dao.InsertWallet(&dao.Wallets{
		Address:      walletAddr,
		EncryptedKey: encryedKeyStr,
	})
	if err != nil {
		return nil, err
	}

	ota, err := utils.StrToBigFloat(req.Originaltokenamount)
	if err != nil {
		return nil, xerrors.New("Invalid origin token amount")
	}

	// 计算输入token价值（可读浮点型字符串）
	var originVal string
	bFloatPrice, _, err := price.MultiTypeChainPrice()
	if err != nil {
		return nil, err
	}

	// 计算U等价值
	switch req.Originaltoken {
	case "ETH":
		val := new(big.Float).Mul(ota, bFloatPrice.EthPerUsdt)
		originVal = fmt.Sprintf("%.18f", val)
		break
	case "USDT":
		originVal = fmt.Sprintf("%.18f", ota)
		break
	case "BTC":
		val := new(big.Float).Mul(ota, bFloatPrice.BtcPerUsdt)
		originVal = fmt.Sprintf("%.18f", val)
		break
	}

	order := &dao.Orders{
		Status:              dao.ORDER_PENDING,
		ReceivedTxInfo:      "",
		ClosedTxInfo:        "",
		Mode:                req.Mode,
		UserReceiveAddress:  req.Userreceiveaddress,
		OriginalToken:       req.Originaltoken,
		OriginalTokenAmount: req.Originaltokenamount,
		OriginalTokenToU:    originVal,
		TargetToken:         req.Targettoken,
		TargetTokenAmount:   req.Targettokenamount,
		WeReceiveAddress:    walletAddr,
		Email:               req.Email,
		Deadline:            time.Now().Add(cnt.DEFAULT_ORDER_DEADLINE * time.Minute),
	}

	// 创建交易订单
	order.Id, err = dao.InsertOrder(order)
	if err != nil {
		return nil, err
	}

	// 发送邮件
	if order.Email != "" {
		go email.SendEmail(order)
	}

	// 推送到队列中 使用 hashOrder
	hashOrder := HashOrder{
		Order:      order,
		InitHeight: nowBlock,
		GasCost:    firstGasCostDef,
	}

	bs, err := hashOrder.MarshalToMsg()
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	err = redis.FirstQueue.PublishBytes(bs)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	// 返回订单信息
	return order, nil
}

func FindOrders(page int, pageSize int) ([]dao.Orders, error) {
	return dao.SelectOrdersByLimit(page, pageSize)
}

func RefundOrder(req *request.RefundReq) (*dao.Orders, error) {
	// 校验邮箱格式
	status := utils.VerifyEmailFormat(req.Email)
	if !status {
		return nil, xerrors.New("Invalid email address")
	}

	orderID, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return nil, err
	}
	order, err := dao.SelectOrderByID(orderID)
	if err != nil {
		return nil, err
	}

	// 判断订单状态是否在 EXPIRED
	if order.Status != dao.ORDER_EXPIRED {
		return nil, xerrors.New("order status not support refund, please contact customer service")
	}

	// 查询Refund表
	refund, err := dao.SelectRefundByOrderID(order.Id)
	if err != nil {
		return nil, err
	}
	if refund.Id != 0 {
		return nil, xerrors.New("Refund records cannot be added repeatedly")
	}

	err = request.MultiWalletCheck(order.OriginalToken, req.Refundaddress)
	if err != nil {
		return nil, err
	}

	decimals18, _ := utils.StrToBigFloat(cnt.DECIMALS_WEI)
	decimalsU, _ := utils.StrToBigFloat(cnt.DECIMALS_USDT)

	// 校验链上子钱包余额
	var leftRefundAmount string
	switch order.OriginalToken {
	case "ETH":
		ethRpc, err := rpc.NewEthRpc()
		if err != nil {
			return nil, xerrors.New("System error: ETH rpc error")
		}

		balance, err := ethRpc.BalanceOfETH(order.WeReceiveAddress)
		if err != nil {
			return nil, err
		}

		bFloat, _ := utils.StrToBigFloat(balance)
		// 判断是否等于0
		if bFloat.Cmp(big.NewFloat(0)) == 0 {
			return nil, xerrors.New("Can not apply for a refund without any balance")
		}

		// 带精度操作
		gasCost, err := ethRpc.GasCost(context.Background(), "ETH")
		gasCostFloat := new(big.Float).SetInt(gasCost)
		bFloat = new(big.Float).Sub(bFloat, gasCostFloat)

		// 转成可读
		leftRefundAmount = new(big.Float).Quo(bFloat, decimals18).String()

		break
	case "USDT":
		ethRpc, err := rpc.NewEthRpc()
		if err != nil {
			return nil, xerrors.New("System error: ETH rpc error")
		}

		balance, err := ethRpc.BalanceOfERC20(order.WeReceiveAddress, rpc.EvmAddrs.UsdtErc20)
		if err != nil {
			return nil, err
		}

		// 返回精度为6 USDT余额
		bFloat, _ := utils.StrToBigFloat(balance)
		// 判断是否等于0
		if bFloat.Cmp(big.NewFloat(0)) == 0 {
			return nil, xerrors.New("Can not apply for a refund without any balance")
		}

		// 转成可读类型
		bFloat = new(big.Float).Quo(bFloat, decimalsU)

		// ETH计价,转成USDT计价
		gasCost, err := ethRpc.GasCost(context.Background(), "ERC20")
		gasCostFloat := new(big.Float).Quo(
			new(big.Float).SetInt(gasCost),
			decimals18,
		)

		readPrice, _, err := price.MultiTypeChainPrice()
		if err != nil {
			return nil, err
		}

		gasCostU := new(big.Float).Mul(gasCostFloat, readPrice.EthPerUsdt)

		leftRefundAmount = new(big.Float).Sub(bFloat, gasCostU).String()

		break
	case "BTC":
		balance, err := btc.GetBalance(order.WeReceiveAddress)
		if err != nil {
			return nil, err
		}

		bFloat, _ := utils.StrToBigFloat(balance)
		// 判断是否等于0
		if bFloat.Cmp(big.NewFloat(0)) == 0 {
			return nil, xerrors.New("Can not apply for a refund without any balance")
		}

		gasCost, _ := utils.StrToBigFloat(cnt.DEFAULT_BTC_GAS)

		leftRefundAmount = new(big.Float).Sub(bFloat, gasCost).String()

		break
	}

	// 更新订单未refund状态
	data := &dao.Refund{
		OrderId:              order.Id,
		ReceiveAddr:          order.UserReceiveAddress,
		PlatformReceivedAddr: order.WeReceiveAddress,
		Email:                req.Email,
		RefundAmount:         leftRefundAmount,
		RefundToken:          order.OriginalToken,
		RefundAddr:           req.Refundaddress,
	}
	_, err = dao.InsertRefund(data)
	if err != nil {
		return nil, err
	}

	session := db.Client().NewSession()
	defer session.Close()
	// 更新订单状态为Refund
	err = dao.UpdateOrderStatusById(
		session,
		order.Id,
		order.Status,
		dao.ORDER_REFUND,
	)

	// 查询最新order数据
	latest, err := dao.SelectOrderByID(order.Id)
	if err != nil {
		return nil, err
	}

	return latest, nil
}

func RefreshOrder(id int64) (*dao.Orders, error) {
	client, err := rpc.NewEthRpc()
	if err != nil {
		return nil, err
	}

	order, err := dao.SelectOrderByID(id)
	if err != nil {
		return nil, err
	}

	if order.Status != dao.ORDER_EXPIRED {
		return nil, xerrors.New("Invalid order ID")
	}

	// 判断流动性
	var leftAmount *big.Float
	tta, _ := utils.StrToBigFloat(order.TargetTokenAmount)
	switch order.TargetToken {
	case "ETH":
		ETHLOCKED, err := redis.GetChainData(constant.GetLqLockedKey(constant.CoinEth))
		if err != nil && !errors.Is(err, orgRedis.Nil) {
			logger.Error("redis GetChainData CoinEth err:", err)
			return nil, err
		}
		ethLocked, _ := utils.StrToBigFloat(ETHLOCKED)

		// 获取钱包eth余额
		ethBalance, err := client.BalanceOfETH(rpc.EthCoreWalletAddr)
		if err != nil {
			return nil, err
		}
		decimals18, _ := utils.StrToBigFloat(cnt.DECIMALS_WEI)
		eBalaceFloat, _ := utils.StrToBigFloat(ethBalance)

		eBalaceFloat = new(big.Float).Quo(eBalaceFloat, decimals18)
		leftAmount = new(big.Float).Sub(eBalaceFloat, ethLocked)
		if leftAmount.Cmp(tta) != 1 {
			return nil, xerrors.New(fmt.Sprintf("%v输出数量不足，重新设置", order.TargetToken))
		}
		break
	case "USDT":
		USDTLOCKED, err := redis.GetChainData(constant.GetLqLockedKey(constant.CoinUsdt))
		if err != nil && !errors.Is(err, orgRedis.Nil) {
			logger.Error("redis GetChainData CoinUsdt err:", err)

			return nil, err
		}

		usdtLocked, _ := utils.StrToBigFloat(USDTLOCKED)

		// 获取钱包eth余额
		UsdtBalance, err := client.BalanceOfERC20(rpc.EthCoreWalletAddr, rpc.EvmAddrs.UsdtErc20)
		if err != nil {
			return nil, err
		}

		decimalsU, _ := utils.StrToBigFloat(cnt.DECIMALS_USDT)
		uBalanceFloat, _ := utils.StrToBigFloat(UsdtBalance)
		uBalanceFloat = new(big.Float).Quo(uBalanceFloat, decimalsU)
		leftAmount = new(big.Float).Sub(uBalanceFloat, usdtLocked)
		if leftAmount.Cmp(tta) != 1 {
			return nil, xerrors.New(fmt.Sprintf("%v输出数量不足，重新设置", order.TargetToken))
		}
		break
	case "BTC":
		BTCLOCKED, err := redis.GetChainData(constant.GetLqLockedKey(constant.CoinBtc))
		if err != nil && !errors.Is(err, orgRedis.Nil) {
			logger.Error("redis GetChainData CoinBtc err:", err)

			return nil, err
		}

		btcLocked, _ := utils.StrToBigFloat(BTCLOCKED)

		btcBalance, err := btc.GetBalance(btc.BtcCoreWallet)
		if err != nil {
			return nil, err
		}
		bBalance, _ := utils.StrToBigFloat(btcBalance)
		leftAmount = new(big.Float).Sub(bBalance, btcLocked)
		if leftAmount.Cmp(tta) != 1 {
			return nil, xerrors.New(fmt.Sprintf("%v输出数量不足，重新设置", order.TargetToken))
		}
		break
	}

	targetTokenAmount := order.TargetTokenAmount
	aAmount, err := utils.StrToBigFloat(targetTokenAmount)
	if err != nil {
		return nil, err
	}

	if leftAmount.Cmp(aAmount) != 1 {
		return nil, xerrors.New("核心钱包余额不足")
	}

	order.Status = dao.ORDER_PENDING
	order.Deadline = time.Now().Add(20 * time.Minute)
	_, err = dao.UpdateOrder(order.Id, dao.ORDER_EXPIRED, order)
	if err != nil {
		return nil, err
	}

	_, firstGasCostDef, err := CalculateOut(
		order.Mode,
		order.OriginalToken,
		order.OriginalTokenAmount,
		order.TargetToken,
	)
	if err != nil {
		return nil, err
	}

	var nowBlock uint64
	if order.OriginalToken == "ETH" || order.OriginalToken == "USDT" {
		// 获取ETH区块高度
		nowBlock, err = redis.GetHeightFormRedis(cnt.ETH_HEIGHT)
		if err != nil {
			return nil, err
		}
	} else if order.OriginalToken == "BTC" {
		nowBlock, err = redis.GetHeightFormRedis(cnt.BTC_HEIGHT)
		if err != nil {
			return nil, err
		}
	}

	// 推送到队列中 使用 hashOrder
	hashOrder := HashOrder{
		Order:      order,
		InitHeight: nowBlock,
		GasCost:    firstGasCostDef,
	}
	bs, err := hashOrder.MarshalToMsg()
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	err = redis.FirstQueue.PublishBytes(bs)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return order, nil
}

func History(page int, pageSize int) ([]HistoryOrder, error) {
	whereField := "status = ?"
	whereVal := "COMPLETED"

	rets, err := dao.SelectOrdersLimitNWhere(page, pageSize, whereField, whereVal)
	if err != nil {
		return nil, err
	}

	var ho []HistoryOrder
	for i := 0; i < len(rets); i++ {
		order := rets[i]
		cmpTime := strings.Split(order.ClosedTxInfo, "*")
		ho = append(ho, HistoryOrder{
			Id:                  order.Id,
			OriginalToken:       order.OriginalToken,
			OriginalTokenAmount: order.OriginalTokenAmount,
			TargetToken:         order.TargetToken,
			TargetTokenAmount:   order.TargetTokenReceived,
			CreatedAt:           order.CreatedAt,
			CompletedAt:         cmpTime[1],
		})
	}

	return ho, nil
}

func Details(id int64) (*dao.Orders, error) {
	data, err := dao.SelectOrderByID(id)
	if err != nil {
		return nil, err
	}
	return data, nil
}

var syncOnce sync.Once

func SyncOrderDataOnceToRedis() {
	logger.Info("===============SyncOrderDataOnceToRedis start===============")
	syncOnce.Do(syncAllCompletedOrderToRedis)
	logger.Info("===============SyncOrderDataOnceToRedis end=================")
}

// syncAllCompletedOrderToRedis 从数据库中同步所有已完成的订单到 redis zset 中, 用于订单金额统计
func syncAllCompletedOrderToRedis() {
	// 从 redis zset 中获取最新的订单
	lastOrder, err := redis.Client().ZRevRange(constant.OrderValueAll, 0, 0).Result()
	if err != nil {
		logger.Error("ZRevRange error: ", err)
		return
	}
	// 如果 redis zset 中没有数据, 则从数据库中同步所有已完成的订单到 redis zset 中
	if len(lastOrder) == 0 {
		syncCompletedOrderToRedisFromDB(0)
		return
	}
	// 如果 redis zset 中有数据, 则判断最新的订单是否已经同步到 redis zset 中
	lastOrderOriginalTokenToU := lastOrder[0]
	// 分隔字符串, 获取订单的 id
	lastOrderOriginalTokenToUSplit := strings.Split(lastOrderOriginalTokenToU, "_")
	if len(lastOrderOriginalTokenToUSplit) != 2 {
		logger.Error("Last order original token to u split error")
		return
	}
	// 获取订单的 id
	lastOrderID, err := strconv.ParseInt(lastOrderOriginalTokenToUSplit[1], 10, 64)
	if err != nil {
		logger.Error("Last order id parse error: ", err)
		return
	}

	// 从数据库中查询最新的订单Count orderId > lastOrderID
	count, err := dao.CountOrderByStatus(db.Client().NewSession(), lastOrderID, dao.ORDER_COMPLETED)
	if err != nil {
		logger.Error("Count completed order error: ", err)
		return
	}
	// 如果数据库中没有最新的订单, 则不需要同步
	if count == 0 {
		logger.Info("No new order, sync to redis exited")
		return
	}
	// 如果数据库中有最新的订单, 则同步到 redis zset 中
	syncCompletedOrderToRedisFromDB(lastOrderID)
}

func syncCompletedOrderToRedisFromDB(gtId int64) {

	// 从数据库中查询所有已完成的订单
	// 分页查询 - 先获取总数
	total, err := dao.CountOrderByStatus(db.Client().NewSession(), gtId, dao.ORDER_COMPLETED)
	if err != nil {
		logger.Error("Count completed order error: ", err)
		return
	}
	// 分页查询
	pageSize := 100
	page := 1
	for {
		if (page-1)*pageSize > int(total) {
			break
		}
		session := db.Client().NewSession().Where("status = ?", "COMPLETED").Where("id > ?", gtId)
		orders, err := dao.SelectOrders(session, page, pageSize)
		if err != nil {
			logger.Error("Select orders error: ", err)
			return
		}
		for _, order := range orders {
			err := redis.Client().ZAdd(constant.OrderValueAll, orgRedis.Z{
				Score:  float64(order.UpdatedAt.Unix()),
				Member: order.OriginalTokenToU + "_" + strconv.FormatInt(order.Id, 10),
			}).Err()
			if err != nil {
				logger.Error("ZAdd error: ", err)
				return
			}
		}
		page++
	}

	return
}

// createSession 创建 session
func createSession(c *gin.Context, orderId int64) error {

	// 生成 session value
	// 读取 已有的 orderIds, 忽略错误 - 可能是第一次创建 session
	redisKey, orderIds, _ := GetSessionValues(sessions.Default(c))
	if orderIds == nil {
		orderIds = make([]int64, 0)
	}
	if redisKey == "" {
		// 生成 redis key - 即 session 的 value
		redisKey = uuid.NewV4().String()
	}
	// 将 orderId 添加到 orderIds
	orderIds = append(orderIds, orderId)
	// 将 orderIds 转为 json
	orderIdsStr, err := json.Marshal(orderIds)
	if err != nil {
		return err
	}
	// 将 orderIds 存入 redis, key 为 redis key - 即 session 的 value
	err = redis.Client().Set(constant.GetSessionKey(redisKey), string(orderIdsStr), constant.SessionExpireHour*time.Hour).Err()
	if err != nil {
		return err
	}
	// 将 redisKey 存入 session
	session := sessions.Default(c)
	session.Set(constant.DefaultOrderSession, redisKey)
	err = session.Save()
	if err != nil {
		return err
	}
	return nil
}

// GetSessionValues 获取 session 中的 orderIds
func GetSessionValues(session sessions.Session) (string, []int64, error) {
	redisKey, ok := session.Get(constant.DefaultOrderSession).(string)
	if !ok || redisKey == "" {
		logger.Errorf("Get session error: %s, %v", redisKey, ok)
		return "", nil, ecode.AccessDenied
	}
	// 从 redis 中获取 orderIds, key 为 orderIdsStr
	sl, err := GetRelatedOrderIds(redisKey)
	return redisKey, sl, err
}

func GetRelatedOrderIds(redisKey string) ([]int64, error) {
	// 从 redis 中获取 orderIds, key 为 orderIdsStr
	orderIdsStr, err := redis.Client().Get(constant.GetSessionKey(redisKey)).Result()
	if err != nil && !errors.Is(err, orgRedis.Nil) {
		return nil, err
	}
	if orderIdsStr == "" {
		return nil, nil
	}
	// 解析 orderIds
	var orderIds []int64
	err = json.Unmarshal([]byte(orderIdsStr), &orderIds)
	if err != nil {
		return nil, ecode.AccessDenied
	}
	return orderIds, nil
}
