package wallet

import (
	"crypto/ecdsa"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"os"
)

func GenKeyEthWallet() (*ecdsa.PrivateKey, *ecdsa.PublicKey, *common.Address, error) {
	pvk, err := crypto.GenerateKey()
	if err != nil {
		return nil, nil, nil, err
	}

	// 生成钱包私钥
	//pData := crypto.FromECDSA(pvk)
	//privateKey := hexutil.Encode(pData)

	// 生成钱包公钥
	//pubData := crypto.FromECDSAPub(&pvk.PublicKey)
	//publicKey := hexutil.Encode(pubData)

	// 生成钱包地址
	addr := crypto.PubkeyToAddress(pvk.PublicKey)

	return pvk, &pvk.PublicKey, &addr, nil
}

// 生成钱包私钥
func GenPrivateKey(pvk *ecdsa.PrivateKey) (string, error) {
	pData := crypto.FromECDSA(pvk)
	privateKey := hexutil.Encode(pData)

	return privateKey, nil
}

// 生成钱包公钥
func GenPublicKey(pvk *ecdsa.PrivateKey) (string, error) {
	pubData := crypto.FromECDSAPub(&pvk.PublicKey)
	publicKey := hexutil.Encode(pubData)

	return publicKey, nil
}

// 生成随机钱包地址
func GenRandomEthWallet() (common.Address, error) {
	var addr common.Address
	pvk, _, _, err := GenKeyEthWallet()
	if err != nil {
		return addr, err
	}

	addr = crypto.PubkeyToAddress(pvk.PublicKey)

	return addr, nil
}

func GenWalletByKey(pvk *ecdsa.PrivateKey) common.Address {
	addr := crypto.PubkeyToAddress(pvk.PublicKey)

	return addr
}

func GenKeyStoreEthWallet(passWord string, path string) (string, error) {
	var filePath string
	keystorePath := path
	if _, err := os.Stat(keystorePath); err != nil {
		err = os.Mkdir(keystorePath, 0777)
		if err != nil {
			return filePath, err
		}
	}

	key := keystore.NewKeyStore(keystorePath, keystore.StandardScryptN, keystore.StandardScryptP)
	account, err := key.NewAccount(passWord)
	if err != nil {
		return filePath, err
	}

	filePath = account.URL.Path

	return filePath, nil
}

func GenPvkObj(pvk string) (*ecdsa.PrivateKey, error) {
	newPrivateKey, err := hexutil.Decode(pvk)
	if err != nil {
		return nil, err
	}

	key, err := crypto.ToECDSA(newPrivateKey)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// BTC钱包地址转换
func ConvertBtcWallet(addr string, isTest bool) (btcutil.Address, error) {
	var params chaincfg.Params
	if isTest {
		params = chaincfg.TestNet3Params
	} else {
		params = chaincfg.MainNetParams
	}

	btcAddr, err := btcutil.DecodeAddress(addr, &params)
	if err != nil {
		return nil, err
	}

	return btcAddr, nil
}
