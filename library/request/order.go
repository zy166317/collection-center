package request

import (
	"collection-center/config"
	"collection-center/library/utils"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/xerrors"
	"math/big"
)

type OrderReq struct {
	Mode                string `json:"mode" form:"mode" binding:"required"`                               // 字符串判断
	Originaltoken       string `json:"originaltoken" form:"originaltoken" binding:"required"`             // origin token name
	Originaltokenamount string `json:"originaltokenamount" form:"originaltokenamount" binding:"required"` // origin token amount
	Targettoken         string `json:"targettoken" form:"targettoken" binding:"required"`                 // target token name
	Targettokenamount   string `json:"targettokenamount" form:"targettokenamount" binding:"required"`
	Userreceiveaddress  string `json:"userreceiveaddress" form:"userreceiveaddress" binding:"required"`
	Email               string `json:"email" form:"email"`
}

type RefundReq struct {
	Id            string `json:"id" form:"id" binding:"required"`
	Refundaddress string `json:"refundaddress" form:"refundaddress" binding:"required"`
	Email         string `json:"email" form:"email"`
}

type RefreshReq struct {
	Id string `json:"id" form:"id" binding:"required"`
}

// 校验token name
func CheckTokenName(name string) (bool, error) {
	if name == "ETH" || name == "USDT" || name == "BTC" {
		return true, nil
	} else {
		return false, xerrors.New("Invalid token name")
	}
}

func CheckOrderMode(mode string) error {
	if mode == "FIXED" || mode == "FLOAT" {
		return nil
	} else {
		return xerrors.New("Invalid token name")
	}
}

// VerifyEvmWalletAddr 校验EVM钱包
func VerifyEvmWalletAddr(walletAddr string) error {
	addr := common.HexToAddress(walletAddr)

	// 检查地址是否有效
	if addr.Hex() == walletAddr {
		return nil
	} else {
		return xerrors.New("Invalid EVM wallet address")
	}
}

// 校验比特币地址
func VerifyBtcWalletAddr(walletAddr string) error {
	// 尝试解析BTC地址
	params := &chaincfg.MainNetParams
	if config.Config().Rpc.Test {
		params = &chaincfg.TestNet3Params
	}
	_, err := btcutil.DecodeAddress(walletAddr, params)
	if err != nil {
		return xerrors.New("Invalid BTC address")
	}

	// 检查地址是否有效
	return nil
}

func VerifyNum(num string) error {
	fNum, err := utils.StrToBigFloat(num)
	if err != nil {
		return err
	}

	if fNum.Cmp(big.NewFloat(0)) != 1 {
		return xerrors.New("Invalid number format")
	}

	return nil
}

func MultiWalletCheck(token string, walletAddr string) error {
	if token == "ETH" || token == "USDT" {
		err := VerifyEvmWalletAddr(walletAddr)
		if err != nil {
			return err
		}
	} else if token == "BTC" {
		err := VerifyBtcWalletAddr(walletAddr)
		if err != nil {
			return err
		}
	} else {
		return xerrors.New("Invalid token")
	}

	return nil
}
