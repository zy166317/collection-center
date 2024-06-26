package config

import (
	"collection-center/internal/btc"
	"collection-center/internal/email"
	"collection-center/internal/logger"
	"collection-center/internal/rpc"
	"collection-center/internal/signClient"
	"collection-center/library/redis"
	"collection-center/service/db"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
	"math/big"
	"mime"
	"os"
)

var (
	env            string
	serverConfig   ServerConfig
	CollectionInfo map[string]*TokenWallet
)

type TokenWallet struct {
	Address   string          // 钱包地址
	Precision decimal.Decimal // 精度
}

type ServerConfig struct {
	Api              *ApiConfig
	Application      *ApplicationConfig
	Log              *logger.LogConfig
	Redis            *redis.RedisConfig
	Database         *db.DBConfig
	Rpc              *Rpc
	CollectionWallet *CollectionWallet
	Hystrix          *HystrixConfig
	Email            *email.EmailConfig
}

type Rpc struct {
	Test           bool
	EthRpc         *rpc.EthRpc
	EthMaxGasPrice int64 //  用于子钱包到主钱包的转账 GasPrice 限制, 单位: wei
	BtcRpc         *btc.BtcRpc
	RemoteSigner   *signClient.RemoteSigner
}

type ApiConfig struct {
	ListenPort         int
	Debug              bool
	LogLevel           string
	Secret             string
	SystemEmail        string
	QueuePrefetchLimit int
	Origin             string // 开发环境跨域 http://192.168.8.63
	JwtSecret          string
}
type ApplicationConfig struct {
	Name string
}

type CollectionWallet struct {
	EthWallet []string
	BtcWallet []string
	SolWallet []string
	TonWallet []string
}

type HystrixConfig struct {
	MaxConcurrent         int
	Timeout               int
	ErrorPercentThreshold int
	SleepWindow           int
}

func Config() *ServerConfig {
	return &serverConfig
}

type OssConfig struct {
	AccessKeyId     string
	AccessKeySecret string
	Endpoint        string
	CommonBucket    string
	RoleArn         string
	RoleSessionName string
}

// SetTestingConfig For testing only
func SetTestingConfig(c *ServerConfig) {
	serverConfig = *c
}

func NewConfig(cctx *cli.Context) (*viper.Viper, error) {
	env = os.Getenv("GO_ENV")
	v := viper.New()
	v.SetConfigType("yaml")
	v.WatchConfig()

	configPath := cctx.String("conf")
	v.AddConfigPath(configPath)
	configName := "config"
	if env != "" {
		configName = "config." + env
	}
	v.SetConfigName(configName)

	err := v.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read the configuration file: %s", err)
	}

	return v, nil
}
func Env() string {
	return env
}

func InitConfig(cctx *cli.Context) {
	//载入配置
	config, err := NewConfig(cctx)
	if err != nil {
		logger.Fatal("Read config error:", err)
	}
	err = config.Unmarshal(&serverConfig)
	if err != nil {
		logger.Fatal("Read config error:", err)
	}
	logger.Info("serverConfig:", serverConfig)
	//载入日志配置
	logger.InitLog(serverConfig.Log.LogPath)
	//载入redis配置
	redis.SetDomain(serverConfig.Api.Origin)
	_, err = redis.SetRedis(serverConfig.Redis)
	if err != nil {
		logger.Fatal("Set redis error:", err)
	}
	//载入db配置
	db.SetDB(serverConfig.Database)
	//载入 rpc 配置信息
	btc.BtcRpcList = *serverConfig.Rpc.BtcRpc
	// remoteSigner
	signClient.SignerConfig = serverConfig.Rpc.RemoteSigner

	//载入 EthMaxGasPrice
	rpc.EthMaxGasPrice = big.NewInt(serverConfig.Rpc.EthMaxGasPrice)

	//载入 evm地址
	//载入核心钱包地址
	rpc.EthRpcUrls = *serverConfig.Rpc.EthRpc

	// 载入等待区块数

	//info, err := dao.GetCollectionInfo()
	//if err != nil {
	//	logger.Fatal("GetCollectionInfo error:", err)
	//}
	//
	//CollectionInfo = map[string]*TokenWallet{}
	//for _, v := range info {
	//	CollectionInfo[v.CoinName] = &TokenWallet{
	//		Address:   v.Address,
	//		Precision: v.Precision,
	//	}
	//}
	////初始化btcd的配置
	//err = btc.InitBtcd(serverConfig.Rpc.Test)
	//if err != nil {
	//	logger.Fatal("InitBtcd error:", err)
	//}
	//
	////载入邮箱配置
	email.InitEmail(serverConfig.Email)
	//
	//ethRpc, err := rpc.NewEthRpc()
	//if err != nil {
	//	logger.Fatal("NewEthRpc error:", err)
	//}
	////获取核心钱包初始化的pending nonce(不使用redis)
	//nonce, err := ethRpc.PendingNonce(common.HexToAddress(serverConfig.CoreWallet.EthWallet), false)
	//if err != nil {
	//	logger.Fatal("PendingNonce error:", err)
	//}
	//
	//err = redis.InitNonceRedis(nonce)
	//if err != nil {
	//	logger.Fatal("redis.InitNonceRedis error:", err)
	//
	//}
	//utils.SetPasswd(serverConfig.Api.Secret)
	/*	// oss
		o := new(OssConfig)
		err = config.Sub("oss").Unmarshal(o)
		if err != nil {
			logger.Fatal("Read oss config error:", err)
		}
		err = oss.SetOSSConfig("", o.AccessKeyId, o.AccessKeySecret, o.Endpoint, o.CommonBucket)

		if err != nil {
			logger.Fatal("Read oss config error:", err)
		}
		oss.RoleArn = o.RoleArn
		oss.RoleSessionName = o.RoleSessionName*/
	//initCollection()
	err = initMimeType()
	if err != nil {
		logger.Fatal("initMimeType error:", err)
	}
}

func initMimeType() error {
	mimes := map[string]string{
		".mp4":       "video/mp4",
		".ogg":       "video/ogg",
		".flv":       "video/flv",
		".avi":       "video/avi",
		".wmv":       "video/wmv",
		".rmvb":      "video/rmvb",
		".mov":       "video/mov",
		".quicktime": "video/quicktime",
	}
	for ext, mimeType := range mimes {
		if mime.TypeByExtension(ext) == "" {
			err := mime.AddExtensionType(ext, mimeType)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//// 加载收款信息
//func initCollection() {
//	info, err := dao.GetCollectionInfo()
//	if err != nil {
//		logger.Fatal("GetCollectionInfo error:", err)
//	}
//	CollectionInfo = map[string]*TokenWallet{}
//	for _, v := range info {
//		CollectionInfo[v.CoinName] = &TokenWallet{
//			Address:   v.Address,
//			Precision: v.Precision,
//		}
//	}
//}
