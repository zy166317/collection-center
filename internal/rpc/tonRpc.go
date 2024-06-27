package rpc

import (
	"collection-center/internal/logger"
	"context"
	"encoding/base64"
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type TonClient struct {
	Client *tonapi.Client
}

func NewTonRpc() (*TonClient, error) {
	//默认主网
	client, err := tonapi.New()
	if err != nil {
		return nil, err
	}
	return &TonClient{
		Client: client,
	}, err
}

// GetTonTransaction 根据MsgID查询交易信息
func (t *TonClient) GetTonTransaction(ctx context.Context, hash string) (*tonapi.Transaction, error) {
	//通过boc解析出msgId
	//msgId, err := ParseBocToMsgID(hash)
	//if err != nil {
	//	return nil, err
	//}
	messageHash, err := t.Client.GetBlockchainTransactionByMessageHash(ctx, tonapi.GetBlockchainTransactionByMessageHashParams{MsgID: hash})
	if err != nil {
		return nil, err
	}
	return messageHash, err
}

func ParseBocToMsgID(boc string) (string, error) {
	decodeString, err := base64.StdEncoding.DecodeString(boc)
	if err != nil {
		logger.Error("GetHashByBoc-DecodeString error:", err)
		return "", err
	}
	fromBOC, err := cell.FromBOC(decodeString)
	if err != nil {
		logger.Error("GetHashByBoc-FromBOC error:", err)
		return "", err
	}
	hash := base64.StdEncoding.EncodeToString(fromBOC.Hash())
	return hash, nil
}

func (t *TonClient) GetWalletBalance(ctx context.Context, address string) error {
	_, err := t.Client.GetAccount(ctx, tonapi.GetAccountParams{AccountID: address})
	if err != nil {
		logger.Error("GetWalletBalance-GetAccount error:", err)
		return err
	}
	return nil
}
