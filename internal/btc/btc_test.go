package btc

import (
	"collection-center/internal/signClient"
	"collection-center/library/redis"
	"collection-center/service/db"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"testing"
)

var testWallet = "tb1quq809m29ggx2t2e0nv0vewzztz78z6u304pfmc"

func init() {
	BtcCoreWallet = "mqdEyyExvwmDyzXRaxSo6Yg8c5yaSKHTL1"
	BtcRpcList.Test = append(BtcRpcList.Test, "go.getblock.io/10aacc88908d47b89997b8f72a180757/")
	BtcRpcList.Mainnet = append(BtcRpcList.Mainnet, "go.getblock.io/0bdcdc1ae4d34e41aebeb1089a821706/")
	signClient.SignerConfig.Host = "192.168.8.63"
	signClient.SignerConfig.Port = "8080"
	signClient.SignerConfig.TlsPemPath = "../../resources/grpc-server-cert.pem"
	signClient.SignerConfig.User = "orca_off_signer"
	signClient.SignerConfig.Pass = "orca_6b9c1bb0f29f80a4fa759d8af2d26dd2"
	InitBtcd(true)
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
	log.Printf("btc init done===========")
	redis.SetRedis(&redis.RedisConfig{
		Addr:         "192.168.8.63:6379",
		Auth:         "orca_redis",
		DialTimeout:  0,
		ReadTimeout:  0,
		WriteTimeout: 0,
		DB:           1,
	})
}

func TestGetTx(t *testing.T) {
	//btcAddr, err := wallet.ConvertBtcWallet("2NFyQxSybFSQV55j8fXNBitczRKKTWvX2Ve", true)
	//if err != nil {
	//	logger.Error(err)
	//}
	//t.Logf("btc address:%v", btcAddr)
	//
	//txs, err := Client.SearchRawTransactions(btcAddr, 0, 100, true, nil)
	//if err != nil {
	//	t.Errorf("Search TX history:%v", err)
	//}
	//
	//t.Logf("TXs:%+v", txs)

	//// 打印交易记录
	//for _, tx := range txs {
	//	fmt.Printf("交易哈希: %s\n", tx.Txid)
	//	fmt.Printf("确认数: %d\n", tx.Confirmations)
	//	fmt.Println("-------------------------------")
	//}

	data, err := GetTxsByAddr("msKM5U5jMpgG6CnCSE5FvwS9Qkbfv93bEi")
	if err != nil {
		t.Error(err)
	}

	t.Logf("Wallet details:%+v", data)

	hash := data.TxRefs[0].TxHash
	t.Logf("Tx hash:%v", hash)
	_, _, gas, _ := GetTxStats(hash)
	t.Logf("Gas:%v", gas)
}

// 测试获取最新高度
func TestGetBlock(t *testing.T) {
	blockCount, err := Client.GetBlockCount()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Block count: %d", blockCount)
}

// GetNetworkInfo rpc节点已废弃
func TestGetNetworkInfo(t *testing.T) {
	networkInfo, err := Client.GetNetworkInfo()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Version: %v", networkInfo.Version)
}
func TestGetUTXOs(t *testing.T) {
	info, err := GetUTXOs("2MwZLumcmHAMuuyT2hGn9psNheHR6tkSvAX")
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("TestGetUTXO: %+v", info)
}

// 获取GetBlockStats
func TestGetblockstats(t *testing.T) {

	info, err := Client.GetBlockStats(2533354, nil)
	if err != nil {
		fmt.Println(err)
	}
	bs, _ := json.Marshal(info)
	fmt.Printf("GetBlockStats: %v", string(bs))
}

func TestGetLatestBlockStats(t *testing.T) {

	info, err := GetLatestBlockStats()
	if err != nil {
		t.Error(err)
	}
	bs, _ := json.Marshal(info)
	fmt.Printf("GetBlockStats: %v", string(bs))
}

func TestEstimatefee(t *testing.T) {
	fee, err := Estimatefee(BtcCoreWallet, testWallet, "0.0002")
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("fee: %v", fee)
}
func TestFormatRawTx(t *testing.T) {
	rawTx, inputs, err := FormatRawTx(testWallet, BtcCoreWallet, "0.0001")
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("TxData: %v,%v", rawTx, inputs)
}
func TestSignRawStringTransaction(t *testing.T) {
	rawTx, inputs, err := FormatRawTx(BtcCoreWallet, testWallet, "0.0002")
	// fmt.Println(rawTx, inputs, err)
	if err != nil {
		t.Error(err)
	}
	signText, err := SignRawStringTransaction_grpc(rawTx, inputs)
	if err != nil {
		t.Error(err)
	} else {
		fmt.Printf("TestSignRawStringTransaction: %v", signText)
	}
}

// 广播签名信息
func TestBroadcastTx(t *testing.T) {
	rawTx := "0200000001fe8439716c94a029195c7d7438cc236e1cb9164013435d1cc6ebad02e8a4a0f6000000006b4830450221009f4d6e7802d7cd7cb10f1b55d1f1610123424c174ae85694a671740e0a467e4d022055758973088cf22fb22d81a29c139176240acc84f66bbc5bed3a583a52daeb420121026f077e91a8bb6ae3cf5f8b00c3ea1b6de58186bd2c427cbea95179913257b359ffffffff011027000000000000160014e00ef2ed45420ca5ab2f9b1eccb84258bc716b9100000000" // Replace with your signed transaction
	txHash, err := BroadcastTx(rawTx)
	if err != nil {
		t.Error(err)
	} else {
		fmt.Printf("Transaction broadcasted with ID: %v", txHash)
	}
}

// 测试发送BTC
func TestSendBTC(t *testing.T) {
	//测试正常发送BTC
	wg := sync.WaitGroup{}
	wg.Add(20)
	list := sync.Map{}
	for i := 0; i < 20; i++ {

		go func() {
			defer wg.Done()
			//time.Sleep(time.Duration(rand.Int31n(20)) * time.Second)
			txHash, err := SendBTC(BtcCoreWallet, testWallet, "0.000001")
			if err != nil {
				t.Error(err)
			} else {
				_, ok := list.Load(txHash)
				if ok {
					panic("++++++++++++++++++++++++++++++++++TestSendBTC: REPEAT!!!" + txHash)
				}
				list.Store(txHash, 1)
				log.Printf("++++++++++++++++++++++++++++++++++TestSendBTC: %v", txHash)
			}
		}()

	}
	wg.Wait()
	fmt.Println(list)
	//
	////测试发送BTC失败 (钱包余额不够)
	//txHash, err2 := SendBTC(BtcCoreWallet, testWallet, "0.000001")
	//if err2.Error() == "insufficient funds" {
	//	log.Printf("Send over amount Btc failed as expect")
	//} else if err2 != nil {
	//	t.Error(err2)
	//} else {
	//	t.Error("Send over amount Btc success as unexpect", txHash)
	//}
}

func TestSendBTCFromChildWallet(t *testing.T) {
	txHash, err := SendBTC(testWallet, BtcCoreWallet, "0.00001")
	if err != nil {
		t.Error(err)
	} else {
		log.Printf("TestSendBTCFromChildWallet: %v", txHash)
	}
}

func TestBulkSendBTC(t *testing.T) {
	for i := 1; i < 10; i++ {
		// time.Sleep(5000 * time.Millisecond)
		amount := strconv.FormatFloat(0.00001*float64(i), 'f', 8, 64)
		txHash, err := SendBTC(BtcCoreWallet, testWallet, amount)
		if err != nil {
			t.Error(err)
		} else {
			log.Printf("TestSendBTC: %v,amount:%v", txHash, amount)
		}
	}
}

func TestGenKeyBTCWallet(t *testing.T) {
	add := GenKeyBTCWallet()
	fmt.Printf("TestGenKeyBTCWallet: %+v", add)
}

func TestGetBalance(t *testing.T) {
	//balance, err := GetBalance(BtcCoreWallet)
	balance, err := GetBalance("msKM5U5jMpgG6CnCSE5FvwS9Qkbfv93bEi")
	if err != nil {
		t.Error(err)
	} else {
		fmt.Printf("TestGenKeyBTCWallet: %+v\n", balance)
	}
}

func TestGetRandomBCYToken(t *testing.T) {
	token := GetRandomBCYToken()
	fmt.Printf("TestGetRandomBCYToken: %+v", token)
}

func TestGetTxStats(t *testing.T) {
	//t.Log(GetTxStats("f9b84d2b1cd6156dc657e3c6864a9d0cee08182346193f89371c8f48953ee7f5"))
	//t.Log(GetTxStats("c64a4c56536633fa6ec26fd3c934601534ea92a2aa0db2e8da13bfeecb723b4e"))
	//t.Log(GetTxStats("c64a4c56536633fa6ec26fd3c934601534ea92a2aa0db2e8da13bfeecb723b41"))

	t.Log(GetTxStats("73e4100eb702fd1b68e7157cba8277af10bffa1423b3a201f79b9eba52212d44"))
}
