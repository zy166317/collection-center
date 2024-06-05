package service

import (
	"collection-center/internal/ecode"
	"collection-center/internal/logger"
	"collection-center/library/constant"
	"collection-center/library/redis"
	"collection-center/library/request"
	"collection-center/library/utils"
	"collection-center/service/db/dao"
	"github.com/gin-gonic/gin"
	orgRedis "github.com/redis/go-redis/v9"
	"golang.org/x/xerrors"
	"math/big"
	"strconv"
	"strings"
	"time"
)

type OrderController struct {
	utils.Controller
}

func NewOrderController(ctx *gin.Context) *OrderController {
	c := &OrderController{}
	c.SetContext(ctx)
	return c
}

func (o *OrderController) TestOrder() {

	err := createSession(o.Ctx, int64(2354343434))
	if err != nil {
		logger.Error("Create session error: ", err)
		o.ResponseErr(err)
	}

	o.ResponseOk(map[string]any{
		"status": true,
	})
}

func (o *OrderController) OrderDetails() {
	queryID := o.Ctx.Query("id")
	id, err := strconv.ParseInt(queryID, 10, 64)
	if err != nil {
		o.ResponseWrapErr(err, ecode.IllegalParam)
		return
	}
	data, err := Details(id)
	if err != nil {
		o.ResponseErr(err)
		return
	}

	type orderDetailExtend struct {
		*dao.Orders
		RefundEmail string `json:"refund_email"`
	}

	resp := orderDetailExtend{
		Orders:      data,
		RefundEmail: "",
	}
	// 如果 data 为 REFUND 状态, 需要获取 refund 的信息, 返回 refund_email
	if data.Status == dao.ORDER_REFUND {
		// 获取 refund 信息
		refund, err := dao.SelectRefundByOrderID(data.Id)
		if err != nil {
			o.ResponseErr(err)
			return
		}
		resp.RefundEmail = refund.Email
	}
	o.ResponseOk(resp)
}

func (o *OrderController) HistoryOrder() {
	data, err := History(1, 10)
	if err != nil {
		o.ResponseErr(err)
		return
	}

	o.ResponseOk(data)
}

func (o *OrderController) GenOrder() {
	req := &request.OrderReq{}
	err := o.Ctx.ShouldBind(req)
	if err != nil {
		o.ResponseErr(err)
		return
	}

	// 字段校验
	err = request.CheckOrderMode(req.Mode)
	if err != nil {
		o.ResponseErr(err)
		return
	}

	err = request.VerifyNum(req.Originaltokenamount)
	if err != nil {
		o.ResponseErr(err)
		return
	}

	err = request.VerifyNum(req.Targettokenamount)
	if err != nil {
		o.ResponseErr(err)
		return
	}

	err = utils.CheckToken(req.Originaltoken)
	if err != nil {
		o.ResponseErr(err)
		return
	}
	err = utils.CheckToken(req.Targettoken)
	if err != nil {
		o.ResponseErr(err)
		return
	}
	if req.Email != "" {
		status := utils.VerifyEmailFormat(req.Email)
		if !status {
			o.ResponseErr(xerrors.New("Invalid email address"))
			return
		}
	}

	err = request.MultiWalletCheck(req.Targettoken, req.Userreceiveaddress)
	if err != nil {
		o.ResponseErr(err)
		return
	}

	order, err := GenerateOrder(req)
	if err != nil {
		o.ResponseErr(err)
		return
	}

	// 生成 session
	err = createSession(o.Ctx, order.Id)
	if err != nil {
		logger.Error("Create session error: ", err)
	}

	o.ResponseOk(order)
}

func (o *OrderController) RefundOrder() {
	req := &request.RefundReq{}
	err := o.Ctx.ShouldBind(req)
	if err != nil {
		o.ResponseErr(err)
		return
	}

	// API参数校验放在RefundOrder内
	data, err := RefundOrder(req)
	if err != nil {
		o.ResponseErr(err)
		return
	}

	o.ResponseOk(data)
}

func (o *OrderController) RefreshOrder() {
	req := &request.RefreshReq{}
	err := o.Ctx.ShouldBind(req)
	if err != nil {
		o.ResponseErr(err)
		return
	}

	num, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		o.ResponseErr(err)
		return
	}

	data, err := RefreshOrder(num)
	if err != nil {
		o.ResponseErr(err)
		return
	}

	o.ResponseOk(data)
}

func (o *OrderController) BriefOrder() {
	/*
		可能存在 redis 传输量过大的情况, 后期可考虑为 订单价值之和 增加缓存 => 即 缓存的缓存
	*/

	// 从 redis 获取 past 24 hours
	past24hours, err := getBriefDataFromRedisToNow(time.Now().Add(-24 * time.Hour))
	if err != nil {
		o.ResponseErr(err)
		return
	}
	// 从 redis 获取 past 7 days 的订单价值之和
	past7days, err := getBriefDataFromRedisToNow(time.Now().Add(-7 * 24 * time.Hour))
	if err != nil {
		o.ResponseErr(err)
		return
	}
	// 从 redis 获取 past 30 days 的订单价值之和
	past30days, err := getBriefDataFromRedisToNow(time.Now().Add(-30 * 24 * time.Hour))
	if err != nil {
		o.ResponseErr(err)
		return
	}
	// 从 redis 获取 all time 的订单价值之和
	allTime, err := getBriefDataFromRedisToNow(time.Now().Add(-1000 * 24 * time.Hour))
	if err != nil {
		o.ResponseErr(err)
		return
	}
	// ['allTime':'200.3','past24':'20.3','past7':'10.3','past30':'30.3']
	o.ResponseOk(map[string]any{
		"allTime": utils.MustToFloat(allTime),
		"past24":  utils.MustToFloat(past24hours),
		"past7":   utils.MustToFloat(past7days),
		"past30":  utils.MustToFloat(past30days),
	})
}

func getBriefDataFromRedisToNow(from time.Time) (toNowValue *big.Float, err error) {
	// redis zset 数据结构
	// 		Score:  float64(order.UpdatedAt.Unix()),
	//		Member: order.OriginalTokenToU + "_" + strconv.FormatInt(order.Id, 10),
	// 从 redis 获取 from => now 时间内的订单价值之和
	nowUnix := time.Now().Unix()
	fromUnix := from.Unix()
	// 从 redis 分页查询
	total, err := redis.Client().ZCount(constant.OrderValueAll, strconv.FormatInt(fromUnix, 10), strconv.FormatInt(nowUnix, 10)).Uint64()
	if err != nil {
		return nil, err
	}
	logger.Debug("total: ", total, " from: ", fromUnix, " to: ", nowUnix)
	if total == 0 {
		// 没有缓存数据, 返回空
		return big.NewFloat(0), nil
	}
	// 从 redis 中获取数据 - 分页查询
	// 使用 total
	pageSize := int64(100)
	page := int64(1)
	pastValueFromRedis := new(big.Float)

	for {
		if (page-1)*pageSize > int64(total) {
			break
		}
		var pastValuesFromRedisList []string
		pastValuesFromRedisList, err = redis.Client().ZRevRangeByScore(constant.OrderValueAll, orgRedis.ZRangeBy{
			Min:    strconv.FormatInt(fromUnix, 10),
			Max:    strconv.FormatInt(nowUnix, 10),
			Count:  pageSize,
			Offset: (page - 1) * pageSize,
		}).Result()
		if err != nil {
			return nil, err
		}
		logger.Debug("pastValuesFromRedisList length: ", len(pastValuesFromRedisList))
		for _, tmp := range pastValuesFromRedisList {
			// 分隔字符串, 获取订单的 id
			v := tmp
			curValue := strings.Split(v, "_")
			if len(curValue) != 2 {
				logger.Error("Last order original token to u split error")
				return
			}
			if curValue[0] == "" {
				logger.Error("Last order original token to u split error: nil value, orderId: ", curValue[1])
				continue
			}
			f, _ := new(big.Float).SetString(curValue[0])
			logger.Debug("orderId: ", curValue[1], " value: ", curValue[0])
			pastValueFromRedis.Add(pastValueFromRedis, f)
		}
		logger.Debug("pastValueFromRedis: ", pastValueFromRedis)
		page++
	}
	return pastValueFromRedis, nil
}
