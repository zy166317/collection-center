package btc

import (
	"bytes"
	"collection-center/internal/logger"
	"collection-center/internal/signClient"
	"collection-center/internal/signClient/pb/offlineSign"
	"collection-center/library/redis"
	"collection-center/library/utils"
	"collection-center/service/db/dao"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
)

var Client *rpcclient.Client

type BtcRpc struct {
	Test    []string `json:"test"`
	Mainnet []string `json:"mainnet"`
}

var BtcRpcList BtcRpc

var BtcCoreWallet string

var WaitBlock int
var WaitBlockCltBtcMax int
var WaitBlockCltBtcMin int

var mnet = &chaincfg.MainNetParams

var mnetStr = "main"

// btcd docs: https://github.com/btcsuite/btcd/tree/master/rpcclient/examples
func InitBtcd(isTest bool) error {

	if len(BtcRpcList.Mainnet) == 0 || len(BtcRpcList.Test) == 0 {
		return errors.New("btc rpc list is empty")
	}

	Host := BtcRpcList.Mainnet[0]
	if isTest {
		mnetStr = "test3"
		mnet = &chaincfg.TestNet3Params
		Host = BtcRpcList.Test[0]
	}

	// Connect to local bitcoin core RPC server using HTTP POST mode.
	connCfg := &rpcclient.ConnConfig{
		Host:         Host,
		User:         "",
		Pass:         "none", //btcd的库必须要有这个参数
		HTTPPostMode: true,   // Bitcoin core only supports HTTP POST mode
		DisableTLS:   false,  // Bitcoin core does not provide TLS by default
	}
	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return err
	}
	Client = client
	return nil
	// defer client.Shutdown()
}

// 获取最新区块的信息，包括改区块的手续费信息,暂时不用
func GetLatestBlockStats() (*btcjson.GetBlockStatsResult, error) {
	blockCount, err := Client.GetBlockCount()
	if err != nil {
		return nil, err
	}
	return Client.GetBlockStats(blockCount, nil)
}

// 获取UTXO列表,暂时不用
func GetUTXOs(address string) (UTXOs, error) {
	url := "https://" + "TODO" + "blockbook/api/utxo/" + address
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	// 读取响应内容
	body, err := ioutil.ReadAll(response.Body)
	var utoxs UTXOs
	err = json.Unmarshal(body, &utoxs)
	if err != nil {
		return nil, err
	}
	return utoxs, err
}

// 生成交易IN的数据格式,暂时不用
func FormatTxIn(from string, amountStr string) (Inputs, error) {
	//获取源地址utxo列表
	utxos, err := GetUTXOs(from)
	if err != nil {
		return nil, err
	}
	var txin Inputs
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return nil, err
	}
	total := float64(0)
	for _, v := range utxos {
		Amount, err := strconv.ParseFloat(v.Amount, 64)
		//txin 需要大于 目标金额的110%，作为手续费
		if total > amount*float64(1.1) {
			break
		} else {
			total += Amount
		}
		if err != nil {
			return nil, err
		}
		txin = append(txin, Input{
			TxId:         v.TxID,
			VOut:         v.VOut,
			WIF:          "",
			RedeemScript: "",
			SegWit:       false,
			Amount:       Amount,
		})
	}
	return txin, nil
}

// 估算交易费用
func Estimatefee(from string, to string, amountStr string) (string, error) {
	logger.Debugf("from: %s,to: %s,amount: %s", from, to, amountStr)
	bodyStruct, err := FormatData(from, to, amountStr)
	if err != nil {
		return "0", err
	}
	if len(bodyStruct.Errors) > 0 {
		return "0", errors.New(bodyStruct.Errors[len(bodyStruct.Errors)-1].Error)
	}
	amountIn := int64(0)
	amountOut := int64(0)
	for _, v := range bodyStruct.Tx.Inputs {
		amountIn += v.OutputValue
	}
	for _, v := range bodyStruct.Tx.Outputs {
		amountOut += v.Value
	}
	x, y := new(big.Float).SetInt64(amountIn-amountOut), big.NewFloat(1e8)
	fee := new(big.Float).Quo(x, y)
	return fee.String(), nil
}

func FormatData(from string, to string, amountStr string) (*BODY, error) {
	// amountFloat, err := strconv.ParseFloat(amountStr, 64)
	amountFloat, err := utils.StrToBigFloat(amountStr)
	if err != nil {
		return nil, err
	}
	amount, _ := new(big.Float).Mul(amountFloat, big.NewFloat(1e8)).Int64()
	url := fmt.Sprintf("https://api.blockcypher.com/v1/btc/%s/txs/new?token=%s", mnetStr, GetRandomBCYToken())
	method := "POST"
	jsonStr := fmt.Sprintf(`{"inputs":[{"addresses": ["%s"]}],"outputs":[{"addresses": ["%s"], "value": %d}]}`, from, to, amount)
	payload := strings.NewReader(jsonStr)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil
	}
	var bodyStruct BODY
	err = json.Unmarshal(body, &bodyStruct)
	if err != nil {
		return nil, err
	}
	return &bodyStruct, nil
}

// 生成交易数据,并格式化成字符串
func FormatRawTx(from string, to string, amountStr string) (string, Inputs, error) {
	// amount, err := strconv.ParseFloat(amountStr, 64)
	bodyStruct, err := FormatData(from, to, amountStr)
	if err != nil {
		return "", nil, err
	}
	WIF := ""
	if from != BtcCoreWallet {
		wlt, err := dao.SelectOrderByAddr(from)
		if err != nil {
			return "", nil, err
		}
		WIF = wlt.EncryptedKey
	}
	var inputs Inputs
	for _, v := range bodyStruct.Tx.Inputs {
		//logger.Info("Get format input tx data from blockcypher,tx hash:", v.PrevHash)
		inputs = append(inputs, Input{
			TxId:         v.PrevHash,
			VOut:         v.OutputIndex,
			WIF:          WIF,
			RedeemScript: "",
			SegWit:       v.ScriptType == "pay-to-witness-pubkey-hash",
			Amount:       float64(v.OutputValue) / 1e8,
		})
	}
	var outputs Outputs
	for _, v := range bodyStruct.Tx.Outputs {
		outputs = append(outputs, Output{
			PayToAddress: v.Addresses[0],
			Amount:       float64(v.Value) / 1e8,
		})
	}
	rawTx, err := CreateRawStringTransaction(inputs, outputs)
	if err != nil {
		logger.Errorf("CreateRawStringTransaction error:%v", err)
		return "", inputs, err
	}
	return rawTx, inputs, nil
}

// SignRawStringTransaction 使用 grpc 远程调用覆盖原有 btcd 包内签名操作
func SignRawStringTransaction_grpc(script string, inputs Inputs) (string, error) {
	client, conn, err := signClient.NewClient()
	if err != nil {
		return "", err
	}
	defer conn.Close()
	btcInputs := make([]*offlineSign.BtcInput, len(inputs))
	for i, v := range inputs {
		input := v
		btcInputs[i] = &offlineSign.BtcInput{
			TxId:         input.TxId,
			VOut:         int64(input.VOut),
			WIF:          input.WIF,
			RedeemScript: input.RedeemScript,
			SegWit:       input.SegWit,
			Amount:       input.Amount,
		}
	}
	resp, err := client.BtcSign(context.Background(), &offlineSign.BtcSignReq{
		BtcInputs: btcInputs,
		Script:    script,
	})
	if err != nil {
		return "", err
	}
	return resp.Signed, nil
}

// 广播交易
func BroadcastTx(rawTx string) (string, error) {
	txBytes, err := hex.DecodeString(rawTx)
	if err != nil {
		return "", err
	}
	tx := wire.NewMsgTx(wire.TxVersion)
	err = tx.Deserialize(bytes.NewReader(txBytes))
	if err != nil {
		return "", err
	}
	txHash, err := Client.SendRawTransaction(tx, true)
	if err != nil {
		return "", err
	} else {
		return txHash.String(), err
	}
}

// SendBTC 发送交易 基于 memPool 特性, txHash 需要做去重处理 - 等五秒
func SendBTC(from string, to string, amount string) (string, error) {

	logger.Debugf("from: %s,to: %s,amount: %s", from, to, amount)
	//重复发送多次，提高成功率
	var err error
	//for i := 0; i < 10; i++ {
	//	if i > 0 {
	//		//第二次重复请求时先等待5s
	//		time.Sleep(5 * time.Second)
	//	}
	rawTx, inputs, err := FormatRawTx(from, to, amount)
	if err != nil {
		return "", err
	}
	var signText string
	if from != BtcCoreWallet {
		signText, err = SignRawStringTransaction_local(rawTx, inputs)
		if err != nil {
			return "", err
		}
	} else {
		signText, err = SignRawStringTransaction_grpc(rawTx, inputs)
		if err != nil {
			return "", err
		}
	}

	// Unique
	if !redis.RateLimitForFunc("SendBTC", utils.GenerateMd5(signText), 24*60*60) {
		logger.Errorf("SendBTC: %s, %s, %s, %s", from, to, amount, "参数相同, 等待下一次发送")
		return "", errors.New("参数相同, 等待下一次发送")
	}

	//构造正则表达式
	re, err := regexp.Compile(`bad-txns-in-belowout, value in \(.*\) < value out \(.*\)`)
	if err != nil {
		return "", err
	}
	txHash, errTemp := BroadcastTx(signText)
	if errTemp != nil {
		if re.MatchString(errTemp.Error()) {
			return "", errors.New("insufficient funds")
		} else {
			return "", errTemp
		}
	} else {
		return txHash, nil
	}
	//}
}

// 生成btc子钱包地址
func GenKeyBTCWallet() Address {
	privateKey := NewPrivateKey()
	wif, _ := btcutil.NewWIF(privateKey, mnet, true)
	addr, _ := ParseWIF(wif.String())
	return *addr
}

// 获取btc钱包地址的余额
func GetBalance(addr string) (string, error) {
	var err error
	//重复发送多次，提高成功率
	for i := 0; i < 20; i++ {
		if i > 0 {
			time.Sleep(1 * time.Second)
		}
		url := fmt.Sprintf("https://api.blockcypher.com/v1/btc/%s/addrs/%s/balance?token=%s", mnetStr, addr, GetRandomBCYToken())
		response, errTemp := http.Get(url)
		if errTemp != nil {
			err = errTemp
			continue
		}
		defer response.Body.Close()
		// 读取响应内容
		body, errTemp := io.ReadAll(response.Body)
		if errTemp != nil {
			err = errTemp
			continue
		}
		var balance BalanceStruct
		err = json.Unmarshal(body, &balance)
		if err != nil {
			err = errTemp
			continue
		}
		x, y := new(big.Float).SetInt64(int64(balance.Balance)), big.NewFloat(1e8)
		balanceBig := new(big.Float).Quo(x, y)
		return balanceBig.String(), err
	}
	return "", err
}

// 获取钱包transaction
func GetTxsByAddr(addr string) (WalletDetails, error) {
	var txs WalletDetails
	var err error
	//重复发送多次，提高成功率
	for i := 0; i < 20; i++ {
		if i > 0 {
			time.Sleep(1 * time.Second)
		}
		url := fmt.Sprintf("https://api.blockcypher.com/v1/btc/%s/addrs/%s", mnetStr, addr)
		response, errTemp := http.Get(url)
		if errTemp != nil {
			err = errTemp
			continue
		}
		defer response.Body.Close()
		// 读取响应内容
		body, errTemp := io.ReadAll(response.Body)
		if errTemp != nil {
			err = errTemp
			continue
		}

		err = json.Unmarshal(body, &txs)
		if err != nil {
			err = errTemp
			continue
		}

		return txs, nil
	}
	return txs, err
}

//func CompareOrderByTx(txs WalletDetails) (string, string, error) {
//	tx, err := FetchTxDetails(txHash)
//	if err != nil {
//		return "", "", err
//	}
//
//}

// GetTxStats
// 获取交易状态
// return bool, bool, error
// 第一个bool表示 区块已经确认
// 第二个bool表示 交易已经进入 memPool ==> 交易已经广播 ==> 子钱包归集到核心钱包的情况下, 认为该交易已经成功
// 第三个 int64 标识 交易费, 单位: 聪
func GetTxStats(txHash string) (bool, bool, int64, error) {
	tx, err := FetchTxDetails(txHash)
	if err != nil {
		return false, false, 0, err
	}

	if tx.Confirmed.Before(time.Now()) && !tx.Confirmed.IsZero() && tx.Received.Before(time.Now()) && !tx.Received.IsZero() {
		return true, true, tx.Fees, nil
	} else if tx.Confirmed.IsZero() && tx.Received.Before(time.Now()) && !tx.Received.IsZero() {
		return false, true, tx.Fees, nil
	}
	return false, false, 0, nil
}

func FetchTxDetails(txHash string) (*TxDetails, error) {
	url := fmt.Sprintf("https://api.blockcypher.com/v1/btc/%s/txs/%s?token=%s", mnetStr, txHash, GetRandomBCYToken())

	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var tx TxDetails
	err = json.Unmarshal(body, &tx)
	if err != nil {
		logger.Errorf("body:%+v", tx)
		return nil, err
	}

	logger.Infof("Tx Details:%+v", tx)

	return &tx, nil
}
