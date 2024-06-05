package redis

import (
	cnt "collection-center/contract/constant"
	"collection-center/internal/logger"
	"collection-center/library/constant"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/adjust/rmq/v5"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"golang.org/x/xerrors"
)

var (
	serviceRedis *RedisClient
	domain       string
)

type RedisConfig struct {
	Addr         string
	Auth         string
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func Client() *RedisClient {
	return serviceRedis
}

func SetDomain(d string) {
	domain = d
}

type RedisClient struct {
	single  *redis.Client
	cluster *redis.ClusterClient
}

func (c *RedisClient) TTL(key string) *redis.DurationCmd {
	if c.single != nil {
		return c.single.TTL(context.Background(), key)
	} else {
		return c.cluster.TTL(context.Background(), key)
	}
}

func (c *RedisClient) Expire(key string, expiration time.Duration) *redis.BoolCmd {
	if c.single != nil {
		return c.single.Expire(context.Background(), key, expiration)
	} else {
		return c.cluster.Expire(context.Background(), key, expiration)
	}
}

func (c *RedisClient) Del(keys ...string) *redis.IntCmd {
	if c.single != nil {
		return c.single.Del(context.Background(), keys...)
	} else {
		return c.cluster.Del(context.Background(), keys...)
	}
}

func (c *RedisClient) Get(key string) *redis.StringCmd {
	if c.single != nil {
		return c.single.Get(context.Background(), key)
	} else {
		return c.cluster.Get(context.Background(), key)
	}
}

func (c *RedisClient) Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	if c.single != nil {
		return c.single.Set(context.Background(), key, value, expiration)
	} else {
		return c.cluster.Set(context.Background(), key, value, expiration)
	}
}

func (c *RedisClient) SetNX(key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	if c.single != nil {
		return c.single.SetNX(context.Background(), key, value, expiration)
	} else {
		return c.cluster.SetNX(context.Background(), key, value, expiration)
	}
}
func (c *RedisClient) SetXX(key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	if c.single != nil {
		return c.single.SetXX(context.Background(), key, value, expiration)
	} else {
		return c.cluster.SetXX(context.Background(), key, value, expiration)
	}
}
func (c *RedisClient) Incr(key string) *redis.IntCmd {
	if c.single != nil {
		return c.single.Incr(context.Background(), key)
	} else {
		return c.cluster.Incr(context.Background(), key)
	}
}

func (c *RedisClient) IncrByFloat(key string, value float64) *redis.FloatCmd {
	if c.single != nil {
		return c.single.IncrByFloat(context.Background(), key, value)
	} else {
		return c.cluster.IncrByFloat(context.Background(), key, value)
	}
}

func (c *RedisClient) Ping() *redis.StatusCmd {
	if c.single != nil {
		return c.single.Ping(context.Background())
	} else {
		return c.cluster.Ping(context.Background())
	}
}

func (c *RedisClient) Watch(txf func(*redis.Tx) error, name string) error {
	if c.single != nil {
		return c.single.Watch(context.Background(), txf, name)
	} else {
		return c.cluster.Watch(context.Background(), txf, name)
	}
}
func (c *RedisClient) BLpop(key ...string) *redis.StringSliceCmd {
	if c.single != nil {
		return c.single.BLPop(context.Background(), 0, key...)
	} else {
		return c.cluster.BLPop(context.Background(), 0, key...)
	}
}

func (c *RedisClient) RPush(key string, value ...interface{}) *redis.IntCmd {
	if c.single != nil {
		return c.single.RPush(context.Background(), key, value...)
	} else {
		return c.cluster.RPush(context.Background(), key, value...)
	}
}
func (c *RedisClient) LPush(key string, value ...interface{}) *redis.IntCmd {
	if c.single != nil {
		return c.single.LPush(context.Background(), key, value...)
	} else {
		return c.cluster.LPush(context.Background(), key, value...)
	}
}
func (c *RedisClient) LIndex(key string, index int64) *redis.StringCmd {
	if c.single != nil {
		return c.single.LIndex(context.Background(), key, index)
	} else {
		return c.cluster.LIndex(context.Background(), key, index)
	}
}
func (c *RedisClient) LRange(key string, start, stop int64) *redis.StringSliceCmd {
	if c.single != nil {
		return c.single.LRange(context.Background(), key, start, stop)
	} else {
		return c.cluster.LRange(context.Background(), key, start, stop)
	}
}
func (c *RedisClient) LTrim(key string, start, stop int64) *redis.StatusCmd {
	if c.single != nil {
		return c.single.LTrim(context.Background(), key, start, stop)
	} else {
		return c.cluster.LTrim(context.Background(), key, start, stop)
	}
}

func SetRedis(c *RedisConfig) (*RedisClient, error) {
	var err error
	addrs := strings.Split(c.Addr, ",")
	if len(addrs) > 1 {
		//集群
		client := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        addrs,
			Password:     c.Auth,
			DialTimeout:  c.DialTimeout,
			ReadTimeout:  c.ReadTimeout,
			WriteTimeout: c.WriteTimeout,
		})
		sessionStore, err = initRedisSession(20, "tcp", addrs[0], c.Auth, strconv.Itoa(c.DB), domain, []byte(constant.TmpPwd))
		if err != nil {
			return nil, err
		}
		serviceRedis = &RedisClient{cluster: client}
		FirstQueue, err = setRedisMq(constant.FirstListenSamWalletQueue, client, nil)
		SecondQueue, err = setRedisMq(constant.SecondListenSamWalletQueue, client, nil)
		ThirdQueue, err = setRedisMq(constant.ThirdListenSamWalletQueue, client, nil)
		CoreToUserQueue, err = setRedisMq(constant.CoreToUserQueue, client, nil)
		if err != nil {
			return nil, err
		}
	} else {
		//非集群
		singleOption := &redis.Options{
			Addr:         c.Addr,
			Password:     c.Auth,
			DB:           c.DB,
			DialTimeout:  c.DialTimeout,
			ReadTimeout:  c.ReadTimeout,
			WriteTimeout: c.WriteTimeout,
		}
		client := redis.NewClient(singleOption)
		serviceRedis = &RedisClient{single: client}
		sessionStore, err = initRedisSession(20, "tcp", c.Addr, c.Auth, strconv.Itoa(c.DB), domain, []byte(constant.TmpPwd))
		if err != nil {
			return nil, err
		}
		FirstQueue, err = setRedisMq(constant.FirstListenSamWalletQueue, nil, singleOption)
		SecondQueue, err = setRedisMq(constant.SecondListenSamWalletQueue, nil, singleOption)
		ThirdQueue, err = setRedisMq(constant.ThirdListenSamWalletQueue, nil, singleOption)
		CoreToUserQueue, err = setRedisMq(constant.CoreToUserQueue, nil, singleOption)
		if err != nil {
			return nil, err
		}
	}
	return serviceRedis, nil
}

func GetLock(lockName string) (string, error) {
	acquireTimeout := constant.AcquireLockTimeout
	lockTimeOut := constant.LockTimeout
	code := uuid.NewString()
	endTime := time.Now().Add(acquireTimeout).UnixNano()
	for time.Now().UnixNano() <= endTime {
		if success, err := Client().SetNX(lockName, code, lockTimeOut).Result(); err != nil && err != redis.Nil {
			return "", err
		} else if success {
			return code, nil
		} else if Client().TTL(lockName).Val() == -1 { //-2:失效；-1：无过期；
			Client().Expire(lockName, lockTimeOut)
		}
		time.Sleep(time.Millisecond * 200)
	}
	return "", errors.New("get redis lock timeout")
}

// var count = 0  // test assist
func ReleaseLock(lockName, code string) bool {
	txf := func(tx *redis.Tx) error {
		if v, err := tx.Get(context.Background(), lockName).Result(); err != nil && err != redis.Nil {
			return err
		} else if v == code {
			_, err := tx.Pipelined(context.Background(), func(pipe redis.Pipeliner) error {
				//count++
				//fmt.Println(count)
				pipe.Del(context.Background(), lockName)
				return nil
			})
			return err
		} else if v != code {
			log.Printf("key -> %s ,redis code-> %s ,not match passed code-> %s\n", lockName, v, code)
			return errors.New("code mismatched!")
		}
		return nil
	}

	for {
		if err := Client().Watch(txf, lockName); err == nil {
			return true
		} else if err == redis.TxFailedErr {
			log.Println("watch key is modified, retry to release lock. err:", err.Error())
		} else {
			log.Println("err:", err.Error())
			return false
		}
	}
}

var FirstQueue rmq.Queue
var SecondQueue rmq.Queue
var ThirdQueue rmq.Queue
var CoreToUserQueue rmq.Queue // 处于 2队列与3队列之间

// setRedisMq 是一个设置队列的示例
// queueName 是队列的名称
// 队列需要在 cmd/Run.go 中进行关闭
// redis.FirstQueue.PublishBytes(bs) <== 入列
func setRedisMq(queueName string, redisClusterClient *redis.ClusterClient, singleOption *redis.Options) (rmq.Queue, error) {
	errChan := make(chan error)
	go func() {
		for {
			logger.Error(<-errChan)
		}
	}()
	var conn rmq.Connection
	var err error
	if redisClusterClient != nil {
		conn, err = rmq.OpenClusterConnection(queueName, redisClusterClient, errChan)
	} else if singleOption != nil {
		conn, err = rmq.OpenConnectionWithRedisOptions(queueName, singleOption, errChan)
	}
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	mq, err := conn.OpenQueue(queueName)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	cleaner := rmq.NewCleaner(conn)
	go func() {
		for {
			returned, err1 := cleaner.Clean()
			if err1 != nil {
				logger.Errorf("rmq.NewCleaner failed to clean: %s, queue: %s", err1, queueName)
			}
			logger.Infof("rmq.NewCleaner cleaned %d, queue: %s", returned, queueName)
			time.Sleep(1 * time.Minute)
		}
	}()
	logger.Infof("init with rmq.NewCleaner done. queue: %s", queueName)

	return mq, nil
}

func InsertChainData(field string, price string) (bool, error) {
	client := Client()

	setCMD := client.Set(field, price, cnt.PRICE_EXPIRED_REDIS)

	setStatus, err := setCMD.Result()
	if err != nil {
		return false, err
	}

	if setStatus != "OK" {
		return false, xerrors.New(fmt.Sprintf("Insert %s failed", field))
	}

	return true, nil
}

func GetChainData(field string) (string, error) {
	ret, err := Client().Get(field).Result()
	if err != nil {
		logger.Error(err)
		return "", err
	}

	return ret, nil
}

func GetHeightFormRedis(heightKey string) (uint64, error) {
	btcHeightStr, err := GetChainData(heightKey)
	if err != nil {
		return 0, err
	}
	nowBlock, err := strconv.ParseInt(btcHeightStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint64(nowBlock), nil
}

func createNonceQueue(value uint64) error {
	//初始化10个nonce
	var data []interface{}
	for i := 0; i < 10; i++ {
		data = append(data, value+uint64(i))
	}
	_, err := Client().RPush(constant.NONCE, data...).Result()
	if err != nil {
		return err
	}
	return nil
}
func InitNonceRedis(value uint64) error {
	//删除旧的数据
	_, err := Client().LTrim(constant.NONCE, 1, 0).Result()
	if err != nil {
		return err
	}
	//初始化10个nonce
	var data []interface{}
	for i := 0; i < 10; i++ {
		data = append(data, value+uint64(i))
	}
	_, err2 := Client().RPush(constant.NONCE, data...).Result()
	if err2 != nil {
		return err2
	}
	logger.Info("InitNonceRedis", value)
	return nil
}

var mutex sync.Mutex

func GetRedisPendingNonce() (uint64, error) {
	mutex.Lock()
	defer mutex.Unlock()
	// 判断队列是否存在
	arr, err := Client().LRange(constant.NONCE, 0, -1).Result()
	if err != nil && err != redis.Nil {
		return 0, err
	}
	if len(arr) == 0 {
		return 0, errors.New("nonce queue is empty")
	}

	v, err := Client().BLpop(constant.NONCE).Result()
	if err != nil {
		return 0, err
	}
	//转为uint
	nonce, err := strconv.ParseUint(v[1], 10, 64)
	if err != nil {
		return 0, err
	}
	logger.Info("GetRedisPendingNonce", nonce)
	arr, err2 := Client().LRange(constant.NONCE, 0, -1).Result()
	if err2 != nil {
		return 0, err2
	}
	if len(arr) < 10 {
		//队列少于10个，生产新的元素
		latestNonce, err := Client().LIndex(constant.NONCE, -1).Uint64()
		if err != nil {
			return 0, err
		}
		_, err1 := Client().RPush(constant.NONCE, latestNonce+1).Result()
		if err1 != nil {
			return 0, err1
		}
	}
	return nonce, err
}

// 余额不够之类导致	momery pool reject 交易的，要把nonce返回到队列 ，这个nonce没上链
// 这里是把nonce返回到队列，下次再来的时候再从队列取nonce
func RejectNonce(value uint64) error {
	mutex.Lock()
	defer mutex.Unlock()
	_, err := Client().LPush(constant.NONCE, value).Result()
	return err
}

func (c *RedisClient) ZAdd(key string, member redis.Z) *redis.IntCmd {
	if c.single != nil {
		return c.single.ZAdd(context.Background(), key, member)
	} else {
		return c.cluster.ZAdd(context.Background(), key, member)
	}
}

func (c *RedisClient) ZIncrBy(key string, member string, value float64) *redis.FloatCmd {
	if c.single != nil {
		return c.single.ZIncrBy(context.Background(), key, value, member)
	} else {
		return c.cluster.ZIncrBy(context.Background(), key, value, member)
	}
}

func (c *RedisClient) ZRem(key string, member string) *redis.IntCmd {
	if c.single != nil {
		return c.single.ZRem(context.Background(), key, member)
	} else {
		return c.cluster.ZRem(context.Background(), key, member)
	}
}

func (c *RedisClient) ZCount(key string, min string, max string) *redis.IntCmd {
	if c.single != nil {
		return c.single.ZCount(context.Background(), key, min, max)
	} else {
		return c.cluster.ZCount(context.Background(), key, min, max)
	}
}

func (c *RedisClient) ZRevRangeByScore(key string, opt redis.ZRangeBy) *redis.StringSliceCmd {
	if c.single != nil {
		return c.single.ZRevRangeByScore(context.Background(), key, &opt)
	} else {
		return c.cluster.ZRevRangeByScore(context.Background(), key, &opt)
	}
}

func (c *RedisClient) ZRevRange(key string, start int64, stop int64) *redis.StringSliceCmd {
	if c.single != nil {
		return c.single.ZRevRange(context.Background(), key, start, stop)
	} else {
		return c.cluster.ZRevRange(context.Background(), key, start, stop)
	}
}

func (c *RedisClient) Exists(key string) bool {
	rst := redis.NewIntResult(0, nil)
	if c.single != nil {
		rst = c.single.Exists(context.Background(), key)
	} else {
		rst = c.cluster.Exists(context.Background(), key)
	}
	return rst.Val() == 1
}
