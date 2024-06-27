package rpc

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type SolClient struct {
	Client *rpc.Client
}

func NewSolRpc() *SolClient {
	client := rpc.New("https://api.devnet.solana.com")
	return &SolClient{
		Client: client,
	}
}

func (s *SolClient) GetSolTransaction(ctx context.Context, hash string) (*rpc.GetTransactionResult, error) {
	signature := solana.MustSignatureFromBase58(hash)
	transaction, err := s.Client.GetTransaction(ctx, signature, nil)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

//// 获取spl token info
//func (s *SolClient) GetTokenInfo(ctx context.Context, tokenAddress string) error {
//	s.Client.GetTokenSupply()
//}
