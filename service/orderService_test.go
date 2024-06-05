package service

import (
	"collection-center/internal/email"
	"collection-center/internal/rpc"
	"collection-center/library/redis"
	"collection-center/library/request"
	"collection-center/service/db"
	"fmt"
	"testing"
)

func init() {
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

	redis.SetRedis(&redis.RedisConfig{
		Addr:         "192.168.8.63:6379",
		Auth:         "orca_redis",
		DialTimeout:  0,
		ReadTimeout:  0,
		WriteTimeout: 0,
	})

	email.InitEmail(&email.EmailConfig{
		Host: "smtp.gmail.com",
		Port: 465,
		User: "orcayihaoji@gmail.com",
		Pass: "ifgmqpidheqgpctm",
	})

	rpc.EthRpcUrls = []string{"https://goerli.infura.io/v3/63b48421bb4e468b935489be18d9dbfc"}

	rpc.EvmAddrs.UsdtErc20 = "0x65E2fe35C30eC218b46266F89847c63c2eDa7Dc7"
	rpc.EvmAddrs.EthGasPriceFeed = "0x169E633A2D1E6c10dD91238Ba11c4A708dfEF37C"
	rpc.EvmAddrs.EthPriceFeed = "0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419"
	rpc.EvmAddrs.BtcPriceFeed = "0xF4030086522a5bEEa4988F8cA5B36dbC97BeE88c"
}

func TestGenerate(t *testing.T) {

	req := &request.OrderReq{
		Mode:                "FIXED",
		Originaltoken:       "ETH",
		Originaltokenamount: "1",
		Targettoken:         "USDT",
		Targettokenamount:   "1570.710973",
		Userreceiveaddress:  "0x22B6e3aBe7F2181D38e9c2F4d3Cd8F15ab1a1bfd",
		Email:               "1535130253@qq.com",
	}
	GenerateOrder(req)
}

func TestSelectOrders(t *testing.T) {
	data, err := FindOrders(1, 5)
	if err != nil {
		t.Fatalf("Select orders error%s\n", err)
	}

	fmt.Printf("Orders:%v\n", data)
}

func TestHistoryOrders(t *testing.T) {
	data, err := History(1, 5)
	if err != nil {
		t.Fatalf("Select orders error%s\n", err)
	}

	fmt.Printf("Orders:%v\n", data)

}

func TestRefund(t *testing.T) {
	req := &request.RefundReq{
		Id:            "1588",
		Refundaddress: "0x22B6e3aBe7F2181D38e9c2F4d3Cd8F15ab1a1bfd",
		Email:         "123123@qq.com",
	}

	data, err := RefundOrder(req)
	if err != nil {
		t.Fatalf("Refund error:%v", err)
	}

	fmt.Printf("Refudn data:%v\n", data)
}
