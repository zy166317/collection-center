package rpc

import (
	"collection-center/internal/logger"
	"encoding/json"
	"fmt"
	"golang.org/x/xerrors"
	"io/ioutil"
	"net/http"
	"time"
)

type TxsArray struct {
	PageSize int                  `json:"page_size"`
	Page     int                  `json:"page"`
	Cursor   string               `json:"cursor"`
	Result   []TransactionDetails `json:"result"`
}

type TransactionDetails struct {
	Hash                     string    `json:"hash"`
	Nonce                    string    `json:"nonce"`
	TransactionIndex         string    `json:"transaction_index"`
	FromAddress              string    `json:"from_address"`
	FromAddressLabel         string    `json:"from_address_label"`
	ToAddress                string    `json:"to_address"`
	ToAddressLabel           string    `json:"to_address_label"`
	Value                    string    `json:"value"`
	Gas                      string    `json:"gas"`
	GasPrice                 string    `json:"gas_price"`
	Input                    string    `json:"input"`
	ReceiptCumulativeGasUsed string    `json:"receipt_cumulative_gas_used"`
	ReceiptGasUsed           string    `json:"receipt_gas_used"`
	ReceiptContractAddress   string    `json:"receipt_contract_address"`
	ReceiptRoot              string    `json:"receipt_root"`
	ReceiptStatus            string    `json:"receipt_status"`
	BlockTimestamp           time.Time `json:"block_timestamp"`
	BlockNumber              string    `json:"block_number"`
	BlockHash                string    `json:"block_hash"`
	TransferIndex            []int     `json:"transfer_index"`
}

func SyncWalletTxs(addr string, network string) ([]TransactionDetails, error) {
	var data TxsArray
	maxRetry := 10
	num := 0

	for {
		if num > maxRetry {
			return nil, xerrors.New("Retry times beyond max limit")
		}

		url := fmt.Sprintf("https://deep-index.moralis.io/api/v2.2/%s?chain=%s", addr, network)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			logger.Errorf("Get transaction details error:%v", err)
			num++

			continue
		}

		// TODO 改成配置文件, 并封装一个随机key调用函数
		apiKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJub25jZSI6IjMzMzU0NDA5LTJmMjItNGE4ZS04YTMzLTVjZTc3YTQxOWFmYiIsIm9yZ0lkIjoiMjE0ODIwIiwidXNlcklkIjoiMjE0NTA2IiwidHlwZUlkIjoiYjBmNTM4MzgtMTEyOC00NGZkLTliZDQtMGRlM2NmZmZlMTUzIiwidHlwZSI6IlBST0pFQ1QiLCJpYXQiOjE2OTk0OTgyOTEsImV4cCI6NDg1NTI1ODI5MX0.0so9O7WYoTRWotElHI-rqhS1yM3As5G57clhOIImGno"

		req.Header.Add("Accept", "application/json")
		req.Header.Add(
			"X-API-Key",
			apiKey,
		)

		res, _ := http.DefaultClient.Do(req)
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			logger.Errorf("Parsing response error:%v", err)
			num++

			continue
		}

		json.Unmarshal(body, &data)

		break
	}

	return data.Result, nil
}

//
//func MatchTxType(txs []TransactionDetails, txType string) (bool, error) {
//	var typeKey string
//	switch txType {
//	case "ETH":
//		typeKey = "0x"
//		break
//	case "ERC20":
//		typeKey = "0xa9059cbb"
//	}
//}
