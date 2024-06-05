package rpc

import (
	"context"
	"github.com/tonkeeper/tonapi-go"
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
	messageHash, err := t.Client.GetBlockchainTransactionByMessageHash(ctx, tonapi.GetBlockchainTransactionByMessageHashParams{MsgID: hash})
	if err != nil {
		return nil, err
	}
	return messageHash, err
}
