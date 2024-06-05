package queue

import (
	"collection-center/internal/btc"
	"collection-center/internal/rpc"
	"collection-center/internal/signClient"
	"collection-center/library/redis"
	"collection-center/service"
	"collection-center/service/db"
	"collection-center/service/db/dao"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"testing"
	"time"
)

func init() {
	btc.BtcCoreWallet = "mqdEyyExvwmDyzXRaxSo6Yg8c5yaSKHTL1"
	btc.BtcRpcList.Test = append(btc.BtcRpcList.Test, "go.getblock.io/10aacc88908d47b89997b8f72a180757")
	btc.BtcRpcList.Mainnet = append(btc.BtcRpcList.Mainnet, "go.getblock.io/0bdcdc1ae4d34e41aebeb1089a821706")
	signClient.SignerConfig.Host = "192.168.8.63"
	signClient.SignerConfig.Port = "8080"
	signClient.SignerConfig.TlsPemPath = "../resources/grpc-server-cert.pem"
	signClient.SignerConfig.User = "orca_off_signer"
	signClient.SignerConfig.Pass = "orca_6b9c1bb0f29f80a4fa759d8af2d26dd2"
	btc.InitBtcd(true)

	db.SetDB(&db.DBConfig{
		DSN:          "orca_web3:orca_web3!@tcp(192.168.8.63:3306)/coinvenni?charset=utf8mb4&parseTime=true&loc=Local",
		ReadDSN:      nil,
		Active:       0,
		Idle:         0,
		ShowSql:      false,
		IdleTimeout:  0,
		QueryTimeout: 0,
		ExecTimeout:  0,
		TranTimeout:  0,
	})

	_, err := redis.SetRedis(&redis.RedisConfig{
		Addr:         "192.168.8.63:6379",
		Auth:         "orca_redis",
		DB:           0,
		DialTimeout:  0,
		ReadTimeout:  0,
		WriteTimeout: 0,
	})
	if err != nil {
		fmt.Printf("-------Redis error:%v\n", err)
		return
	}

	rpc.EthRpcUrls = []string{"https://eth-goerli.g.alchemy.com/v2/ALmhXu_g7MrNqg9bB5TSZj0Ocxv6X0Iq"}

	rpc.EvmAddrs.UsdtErc20 = "0x65E2fe35C30eC218b46266F89847c63c2eDa7Dc7"
	rpc.EvmAddrs.EthGasPriceFeed = "0x169E633A2D1E6c10dD91238Ba11c4A708dfEF37C"
	rpc.EvmAddrs.EthPriceFeed = "0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419"
	rpc.EvmAddrs.BtcPriceFeed = "0xF4030086522a5bEEa4988F8cA5B36dbC97BeE88c"
}

func TestNilBigfloat(t *testing.T) {
	type TestFloat struct {
		Num *big.Float
	}

	data := TestFloat{}

	json.Marshal(data)

	num := new(big.Float).Mul(data.Num, big.NewFloat(100))
	fmt.Printf("Test zero big.float value:%.18f\n", num)
}

func TestSyncReceivedTx(t *testing.T) {
	//hash, gasCost, err := syncReceivedTx("msKM5U5jMpgG6CnCSE5FvwS9Qkbfv93bEi", "BTC", "0.00123", 2537637)
	//if err != nil {
	//	t.Errorf("Sync error:%s", err)
	//}
	//
	//t.Logf("Tx hash:%v", hash)
	//t.Logf("Tx gasCost:%v", gasCost)

	hash, gasCost, err := syncReceivedTx("0x9FA01978B11cDB276df98550789a2E6C7a9219CF", "ETH", "0.1", 0)
	if err != nil {
		t.Errorf("Sync error:%s", err)
	}

	t.Logf("Tx hash:%v", hash)
	t.Logf("Tx gasCost:%v", gasCost)
}

func TestSyncEthGasFee(t *testing.T) {
	gas, err := SyncEthGasFee("0x2d7283665f645676d2a68a905e747c473050852be19d25aef694d89a07b972d1")
	if err != nil {
		t.Fatalf("Sync eth gas fee error:%v", err)
	}

	fmt.Printf("Gas fee:%.18f ETH\n", gas)
}

func TestSyncGasFee(t *testing.T) {
	t.Log(SyncEthGasFee("0x395b2d69106f7560d88719b673e5ffcf8ba37158b299a48b54af541a390bc33d"))
}

func TestRedisFuncLock(t *testing.T) {
	for i := 0; i < 12; i++ {
		t.Log(redis.RateLimitForFunc("test", "test", 10))
		t.Log(redis.RateLimitForFunc("test", strconv.Itoa(i), 10))
		time.Sleep(time.Second)
	}
}

func TestLeftAmount(t *testing.T) {
	//ethRpc, _ := rpc.NewEthRpc()

	data := service.HashOrder{
		Order: &dao.Orders{
			Id:                  0,
			Status:              "",
			ReceivedTxInfo:      "",
			ClosedTxInfo:        "",
			Mode:                "",
			UserReceiveAddress:  "0x22B6e3aBe7F2181D38e9c2F4d3Cd8F15ab1a1bfd",
			OriginalToken:       "USDT",
			OriginalTokenAmount: "10000",
			TargetToken:         "BTC",
			TargetTokenAmount:   "0.2662135771",
			TargetTokenReceived: "",
			WeReceiveAddress:    "2NFyQxSybFSQV55j8fXNBitczRKKTWvX2Ve",
		},
		CollectedHeight: 2534452,
		CollectedHash:   "5ca752ee58d67edfc82f299dd41f131a6b980c37a6f08cbc27a9ce8459569b31",
		GasCost:         nil,
		SendHash:        "",
		SendHeight:      0,
	}

	fmt.Printf("data:%+v\n", data)

	leftAmount, _, err := calculateLeftAmount(data, data.Order.TargetToken)
	if err != nil {
		t.Error(err)
	}
	t.Logf("leftamount:%v", leftAmount)

	//decimals18, _ := utils.StrToBigFloat(cnt.DECIMALS_WEI)
	//lInt := new(big.Int)
	//new(big.Float).Mul(leftAmount, decimals18).Int(lInt)
	//
	//logger.Infof("发送target amount:%v", lInt)

	//receiver := common.HexToAddress(data.Order.UserReceiveAddress)
	//tx, _, err := ethRpc.SendEthOffSign(lInt, receiver)
	//if err != nil {
	//	t.Error(err)
	//}
	//
	//logger.Warnf("SendEthOffSign[target-ETH]:%v\n", tx.Hex())
}
