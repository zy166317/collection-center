package utils

import (
	"collection-center/library/constant"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gagliardetto/solana-go"
	"github.com/tonkeeper/tongo"
)

func CheckWalletAddress(address map[string]string) error {
	for k, v := range address {
		if k == constant.EthChain {
			if b := common.IsHexAddress(v); !b {
				return fmt.Errorf("eth address error")
			}
		}
		if k == constant.TonChain {
			_, err := tongo.ParseAddress(v)
			if err != nil {
				return fmt.Errorf("ton address error")
			}
		}
		if k == constant.SolChain {
			out, err := solana.PublicKeyFromBase58(v)
			if err != nil {
				return fmt.Errorf("sol address error")
			}
			if !out.IsOnCurve() {
				return fmt.Errorf("sol address error")
			}
		}
	}
	return nil
}
