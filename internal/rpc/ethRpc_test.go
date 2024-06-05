package rpc

import (
	"collection-center/internal/signClient"
	"collection-center/library/redis"
	"collection-center/library/wallet"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func init() {
	signClient.SignerConfig = &signClient.RemoteSigner{
		Host:       "192.168.8.63",
		Port:       "8080",
		TlsPemPath: "../../resources/grpc-server-cert.pem",
		User:       "orca_off_signer",
		Pass:       "orca_6b9c1bb0f29f80a4fa759d8af2d26dd2",
	}
	redis.SetRedis(&redis.RedisConfig{
		Addr:         "192.168.8.63:6379",
		Auth:         "orca_redis",
		DialTimeout:  0,
		ReadTimeout:  0,
		WriteTimeout: 0,
		DB:           1,
	})

	//EthRpcUrls = []string{"https://goerli.infura.io/v3/63b48421bb4e468b935489be18d9dbfc"}
	EthRpcUrls = []string{"https://eth-goerli.g.alchemy.com/v2/Vo7tbYU-XxlEwjpIVfzNYMndZGTIfE9V"}

	EvmAddrs.UsdtErc20 = "0x65E2fe35C30eC218b46266F89847c63c2eDa7Dc7"
	EvmAddrs.EthGasPriceFeed = "0x169E633A2D1E6c10dD91238Ba11c4A708dfEF37C"
	EvmAddrs.EthPriceFeed = "0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419"
	EvmAddrs.BtcPriceFeed = "0xF4030086522a5bEEa4988F8cA5B36dbC97BeE88c"

	redis.InitNonceRedis(4935)

}

func TestTxHash(t *testing.T) {
	client, err := NewEthRpc()
	if err != nil {
		t.Errorf("Connect error:%v", err)
	}
	hash := common.HexToHash("0x56765d86e2fb1f6c66f5d5fbb0d5474731e298c1e21bc3582ec08c768ecb1d23")
	details, isPending, err := client.Client.TransactionByHash(context.Background(), hash)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("Details:%+v\n", details)

	fmt.Printf("isPending:%v\n", isPending)
	fmt.Printf("details:%+v\n", details.Time())
	fmt.Printf("gas:%v\n", details.Gas())
	fmt.Printf("cost:%v\n", details.Cost())
}

func TestReceipt(t *testing.T) {
	client, err := NewEthRpc()
	if err != nil {
		t.Errorf("Connect error:%v", err)
	}

	hash := common.HexToHash("0xc02401b3210ec575a7fd5f8f8375e1088a0018341a74cf67db44678e0b3af82f")
	receipt, err := client.SyncTxReceipt(context.Background(), &hash)
	if err != nil {
		t.Errorf("Sync error:%v\n", err)
	}

	t.Logf("Receipt:%+v", receipt)
}

func TestConnectEthRPC(t *testing.T) {
	client, err := NewEthRpc()
	if err != nil {
		t.Errorf("Connect error:%v", err)
	}

	//t.Logf("ChainID:%d", chainId)
	fmt.Printf("ChainID:%d\n", client.ChainID)
	fmt.Printf("RPC url:%s\n", client.RpcUrl)
	fmt.Printf("Network:%s\n", client.Network)
}

// 远程调用离线转账
func TestSignTx(t *testing.T) {
	ethRpc, err := NewEthRpc()
	if err != nil {
		t.Error(err)
		return
	}

	//amount, _ := utils.StrToBigInt(cnt.DECIMALS_WEI)
	amount := big.NewInt(1000000000000000000) // in wei (0.001 ETH)
	to := common.HexToAddress("0x22B6e3aBe7F2181D38e9c2F4d3Cd8F15ab1a1bfd")
	tx, _, err := ethRpc.SendEthOffSign(amount, to)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("Hash:%v\n", tx.Hex())
}

func TestSendETH(t *testing.T) {
	newPrivateKey, _ := hexutil.Decode("0x" + "17ce9e0188c492eb86a7de6d3e295fdefbaeff76502f928fe7913b1b76610e02")
	pvk, _ := crypto.ToECDSA(newPrivateKey)
	ethRpc, err := NewEthRpc()
	if err != nil {
		t.Errorf("Send ETH NewEthRpc error:%T,%+v", err, err)
	}
	// amount, _ := new(big.Int).SetString("1000000000", 10)
	to := common.HexToAddress("0x3535271abC21B4ce48442bB504D83Cdab4943B7e")
	fromAddress := wallet.GenWalletByKey(pvk)
	nonce, err := ethRpc.PendingNonce(fromAddress)
	// uint64()
	//  int转换为uint64
	nonce = nonce + uint64(rand.Intn(2))
	fmt.Printf("nonce: %v\n", nonce)
	if err != nil {
		t.Error("PendingNonce error:", err)
	}
	balance, _ := ethRpc.BalanceOfETH("0x9423f2BC63004aa25eEe7aB64663B16Ef658F2C5")
	balance2, _ := new(big.Int).SetString(balance[0:len(balance)-1], 10)
	amount := new(big.Int).Mul(balance2, big.NewInt(int64(rand.Intn(5)+5)))
	fmt.Printf("balance: %v\n", balance)
	fmt.Printf("amount: %v\n", amount)
	/* if i == 1 {
		amount = new(big.Int).Mul(balance2, big.NewInt(7))
		nonce = nonce + 1
	} */
	hash, err := ethRpc.SendETH(SendingInfo{
		PvKey:    pvk,
		Amount:   amount,
		Receiver: to,
	}, nonce)
	if err != nil {
		t.Errorf("Send ETH error:%v\n", err)
	} else {
		fmt.Printf("Send eth hash:%v\n", hash)
	}

}

func TestGoAndFor(t *testing.T) {
	for i := 0; i < 10; i++ {
		go func() {
			fmt.Println(i)
		}()
	}
	time.Sleep(3 * time.Second)
}

func TestSendETHWrapper(t *testing.T) {
	//newPrivateKey, _ := hexutil.Decode("0x" + "2f219ebf353f8c3f5c3cd691d03b92356b9c1bb0f29f80a4fa759d8af2d26dd2")
	//pvk, _ := crypto.ToECDSA(newPrivateKey)
	//ethRpc, _ := NewEthRpc()
	//ethRpc.SetEthCoreWallet("0x6100245165563350748684540982952184534421")
	//to := common.HexToAddress("0xa3c659e7384aA1BAa4832Ea7b616600661de22f3")
	//amount, _ := new(big.Int).SetString("1000000000000000000000000000000000", 10)
	//// amount, err := big.NewInt().SetString("0.001 * 1e18") // in wei (0.001 ETH)
	//hash, err := ethRpc.SendTx(SendingInfo{
	//	PvKey:    pvk,
	//	Amount:   amount,
	//	Receiver: to,
	//})
	//if err != nil {
	//	t.Errorf("Send ETH error:%v\n", err)
	//} else {
	//	fmt.Printf("Send eth hash:%v\n", hash)
	//}
}

func TestSendERC20(t *testing.T) {
	newPrivateKey, _ := hexutil.Decode("0x" + "4c5bdc1bc2d213e6b67f23132dc7767a6d877183d6e435b1711a2d72b66d106e")
	pvk, _ := crypto.ToECDSA(newPrivateKey)
	ethRpc, _ := NewEthRpc()

	amount := big.NewInt(1000000)
	to := common.HexToAddress("0xa3c659e7384aA1BAa4832Ea7b616600661de22f3")
	fromAddress := wallet.GenWalletByKey(pvk)
	nonce, err := ethRpc.PendingNonce(fromAddress)
	if err != nil {
		t.Errorf("PendingNonce error:%v\n", err)
	}
	hash, err := ethRpc.SendERC20(SendingInfo{
		PvKey:     pvk,
		Amount:    amount,
		Receiver:  to,
		TokenAddr: "0x65E2fe35C30eC218b46266F89847c63c2eDa7Dc7",
	}, nonce)
	if err != nil {
		t.Errorf("Send TOKEN error:%v\n", err)
	}

	fmt.Printf("Send token hash:%v\n", hash)
}
func TestSendERC20Wrapper(t *testing.T) {
	newPrivateKey, _ := hexutil.Decode("0x" + "4c5bdc1bc2d213e6b67f23132dc7767a6d877183d6e435b1711a2d72b66d106e")
	pvk, _ := crypto.ToECDSA(newPrivateKey)
	ethRpc, _ := NewEthRpc()
	amount, _ := new(big.Int).SetString("1000000000000000000000000000000000000000000000000000000000000000000000000000000", 10)
	// amount := big.NewInt(1000000)
	to := common.HexToAddress("0xa3c659e7384aA1BAa4832Ea7b616600661de22f3")
	hash, err := ethRpc.SendTx(SendingInfo{
		PvKey:     pvk,
		Amount:    amount,
		Receiver:  to,
		TokenAddr: "0x65E2fe35C30eC218b46266F89847c63c2eDa7Dc7",
	})
	if err != nil {
		t.Errorf("Send TOKEN error:%v\n", err)
	} else {
		fmt.Printf("Send token hash:%v\n", hash)
	}

}
func TestSendERC20OffSign(t *testing.T) {
	ethRpc, _ := NewEthRpc()

	amount := big.NewInt(1000000)
	to := common.HexToAddress("0xa3c659e7384aA1BAa4832Ea7b616600661de22f3")

	hash, _, err := ethRpc.SendERC20OffSign(amount, to, "0x65E2fe35C30eC218b46266F89847c63c2eDa7Dc7")
	if err != nil {
		t.Errorf("Send TOKEN error:%v\n", err)
	}

	fmt.Printf("Send token hash:%v\n", hash)
}

func TestHashRecipt(t *testing.T) {
	newPrivateKey, _ := hexutil.Decode("0x" + "4c5bdc1bc2d213e6b67f23132dc7767a6d877183d6e435b1711a2d72b66d106e")
	pvk, _ := crypto.ToECDSA(newPrivateKey)
	ethRpc, _ := NewEthRpc()

	amount := big.NewInt(1000000)
	to := common.HexToAddress("0xa3c659e7384aA1BAa4832Ea7b616600661de22f3")
	fromAddress := wallet.GenWalletByKey(pvk)
	nonce, err := ethRpc.PendingNonce(fromAddress)
	if err != nil {
		t.Errorf("PendingNonce error:%v\n", err)
	}
	hash, err := ethRpc.SendERC20(SendingInfo{
		PvKey:     pvk,
		Amount:    amount,
		Receiver:  to,
		TokenAddr: "0x65E2fe35C30eC218b46266F89847c63c2eDa7Dc7",
	}, nonce)
	if err != nil {
		t.Errorf("Send TOKEN error:%v\n", err)
	}
	fmt.Printf("Send token hash:%v\n", hash.String())

	fmt.Println("Waiting 15 seconds")
	time.Sleep(15 * time.Second)

	// Check hash status
	receipt, err := ethRpc.SyncTxReceipt(context.Background(), hash)
	if err != nil {
		t.Errorf("Sync error:%v\n", err)
	}

	fmt.Printf("Hash:%v\n", receipt.TxHash)
	fmt.Printf("Status:%d\n", receipt.Status)
}

func TestETHBalance(t *testing.T) {
	ethRpc, _ := NewEthRpc()
	b, err := ethRpc.BalanceOfETH("0x9423f2BC63004aa25eEe7aB64663B16Ef658F2C5")
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("ETH balance:%s\n", b)
}

func TestERC20Balance(t *testing.T) {
	ethRpc, _ := NewEthRpc()
	b, err := ethRpc.BalanceOfERC20("0x997356329e962E1b7ff3fAc63939D7c0973657Df", "0x65E2fe35C30eC218b46266F89847c63c2eDa7Dc7")
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("Token balance:%s\n", b)
}

func TestPendingNonce(t *testing.T) {
	//ethRpc, _ := NewEthRpc()
	//ethRpc.SetEthCoreWallet("0x6100245165563350748684540982952184534421")
	//for i := 0; i < 100; i++ {
	//	go func() {
	//		nonce, err := ethRpc.PendingNonce(common.HexToAddress("0x6100245165563350748684540982952184534421"))
	//		if err != nil {
	//			t.Error(err)
	//		}
	//		fmt.Printf("nonce: %v\n", nonce)
	//	}()
	//}
	//time.Sleep(10 * time.Second)
}

// 测试nonce队列
func TestGetRedisPendingNonce(t *testing.T) {
	err := redis.InitNonceRedis(120)
	if err != nil {
		t.Error(err)
		return
	}
	nonce, err := redis.GetRedisPendingNonce()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("nonce r: %v\n", nonce)
}

func TestMainnetGasPrice(t *testing.T) {
	ctx := context.Background()
	url := "https://ethereum.publicnode.com"

	client, err := ethclient.DialContext(ctx, url)
	if err != nil {

		t.Error(err)
	}

	gas, err := client.SuggestGasPrice(ctx)
	if err != nil {
		t.Error(err)
	}

	t.Logf("Mainnet gas price:%v", gas)

	clientTest, err := NewEthRpc()
	if err != nil {
		t.Error(err)
	}
	gas, err = clientTest.Client.SuggestGasPrice(context.Background())
	if err != nil {
		t.Error(err)
	}
	t.Logf("Testnet gas price:%v", gas)

}

// 测试 gas
func TestTwoDiffGas(t *testing.T) {
	client, err := NewEthRpc()
	if err != nil {
		t.Errorf("Connect error:%v", err)
		return
	}
	gas, err := client.Client.SuggestGasPrice(context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("gas:%v\n", gas)
	gas2, err := client.Client.SuggestGasTipCap(context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("gas2:%v\n", gas2)
}

func TestGetAddrTransfers(t *testing.T) {
	client, err := NewEthRpc()
	if err != nil {
		t.Errorf("Connect error:%v", err)
		return
	}
	t.Log(client.GetAddrTransfers("0x00dCB7E9C07F7C87b5fc429a132caaaCada27ff3", 0, "ETH", "0.0000000003066375"))
	//t.Log(client.GetAddrTransfers("0x8b066b2c4cD0d5957ac3Cc21aE9DfF031191b95A", 0, "USDT", "500"))
	// hash, gasFee - 单位 eth , err
}

func TestGasPrice(t *testing.T) {
	client, err := NewEthRpc()
	if err != nil {
		t.Errorf("Connect error:%v", err)
		return
	}
	t.Log(client.SuggestGasPrice(context.Background()))

	client, err = NewEthRpc(true)
	if err != nil {
		t.Errorf("Connect error:%v", err)
		return
	}
	t.Log(client.SuggestGasPrice(context.Background()))
}

func TestQueryTx(t *testing.T) {
	addr := "0x8b066b2c4cD0d5957ac3Cc21aE9DfF031191b95A"
	netWork := "goerli"

	data, err := SyncWalletTxs(addr, netWork)
	if err != nil {
		t.Error(err)
	}

	t.Logf("data:%+v", data)
	t.Logf("hash:%v", data[0].Hash)
	t.Logf("Input:%v", data[0].Input)
}
