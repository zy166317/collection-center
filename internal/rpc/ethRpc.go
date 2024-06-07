package rpc

import (
	"collection-center/contract/build"
	cnt "collection-center/contract/constant"
	"collection-center/internal/logger"
	"collection-center/internal/signClient"
	"collection-center/internal/signClient/pb/offlineSign"
	"collection-center/library/constant"
	"collection-center/library/redis"
	"collection-center/library/utils"
	"collection-center/library/wallet"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/go-errors/errors"
	"github.com/shopspring/decimal"
	"golang.org/x/xerrors"
	"math/big"
	"reflect"
	"strings"
	"time"
)

type EthClient struct {
	Client  *ethclient.Client
	ChainID *big.Int
	Network string
	RpcUrl  string
}

// 原生ETH RPC
type BaseEthClient struct {
	Client *rpc.Client
}

type BaseTx struct {
	AccessList           []interface{} `json:"accessList"`
	BlockHash            string        `json:"blockHash"`
	BlockNumber          string        `json:"blockNumber"`
	ChainId              string        `json:"chainId"`
	From                 string        `json:"from"`
	Gas                  string        `json:"gas"`
	GasPrice             string        `json:"gasPrice"`
	Hash                 string        `json:"hash"`
	Input                string        `json:"input"`
	MaxFeePerGas         string        `json:"maxFeePerGas"`
	MaxPriorityFeePerGas string        `json:"maxPriorityFeePerGas"`
	Nonce                string        `json:"nonce"`
	R                    string        `json:"r"`
	S                    string        `json:"s"`
	To                   string        `json:"to"`
	TransactionIndex     string        `json:"transactionIndex"`
	Type                 string        `json:"type"`
	V                    string        `json:"v"`
	Value                string        `json:"value"`
	YParity              string        `json:"yParity"`
}

type EthRpc []string

type EvmAddress struct {
	UsdtErc20       string
	EthPriceFeed    string
	EthGasPriceFeed string
	BtcPriceFeed    string
}
type SendingInfo struct {
	PvKey     *ecdsa.PrivateKey
	Amount    *big.Int
	Receiver  common.Address
	TokenAddr string
}

type EthTransfer struct {
	Hash    string
	From    string
	To      string
	Amount  decimal.Decimal
	OrderId string
}

var (
	EthRpcUrls         []string
	EvmAddrs           EvmAddress
	EthCoreWalletAddr  string
	WaitBlock          int
	WaitBlockCltEthMax int
	WaitBlockCltEthMin int
	EthMaxGasPrice     *big.Int
)

func NewEthRpc(usingMainnet ...bool) (*EthClient, error) {
	ctx := context.Background()
	//url := "https://ethereum.publicnode.com"

	// 随机生成rpc url
	url, _, err := utils.RandomEthRpcUrl(EthRpcUrls)
	if err != nil {
		return nil, err
	}

	if len(usingMainnet) > 0 && usingMainnet[0] {
		url = strings.Replace(url, "goerli", "mainnet", 1)
		logger.Debug("using mainnet url:", url)
	}

	client, err := ethclient.DialContext(ctx, url)
	if err != nil {
		return nil, err
	}
	chainId, err := client.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	return &EthClient{
		Client:  client,
		ChainID: chainId,
		Network: utils.MatchNetwork(chainId),
		RpcUrl:  url,
	}, nil
}

func NewBaseEthRpc(usingMainnet ...bool) (*BaseEthClient, error) {
	ctx := context.Background()
	//url := "https://ethereum.publicnode.com"

	// 随机生成rpc url
	url, _, err := utils.RandomEthRpcUrl(EthRpcUrls)
	if err != nil {
		return nil, err
	}

	if len(usingMainnet) > 0 && usingMainnet[0] {
		url = strings.Replace(url, "goerli", "mainnet", 1)
		logger.Debug("using mainnet url:", url)
	}
	client, err := rpc.DialContext(ctx, url)
	if err != nil {
		return nil, err
	}
	return &BaseEthClient{Client: client}, nil
}

func (e *EthClient) Close() {
	e.Client.Close()
}

// PendingNonce 获取eth地址的pending nonce
//
//	receiver e EthClient
//	param account eth地址
//	param redisCache 是否使用redis缓存，只有当地址是核心钱包地址才能生效，默认true
//	return uint64 nonce
//	return error
func (e *EthClient) PendingNonce(account common.Address, redisCache ...bool) (uint64, error) {
	redisCacheFlag := true
	if len(redisCache) > 0 {
		redisCacheFlag = redisCache[0]
	}
	logger.Debug("------------- redisCacheFlag: ", redisCacheFlag, " redisCache:", redisCache, " account: ", account.Hex(), " EthCoreWalletAddr: ", EthCoreWalletAddr)
	if EthCoreWalletAddr == account.Hex() && redisCacheFlag {
		logger.Debug("+++++++++++++PendingNonce use redis queue: account:", account.Hex())
		nonce, err := redis.GetRedisPendingNonce()
		return nonce, err
	}
	logger.Debug("--------------PendingNonce use Client: account:", account.Hex())
	nonce, err := e.Client.PendingNonceAt(context.Background(), account)
	if err != nil {
		return 0, err
	}

	return nonce, nil
}

func (e *EthClient) SendETH(sendingInfo SendingInfo, nonce uint64) (*common.Hash, error) {
	ctx := context.Background()

	signedTx, err := e.SendEthTransaction(ctx, sendingInfo.PvKey, sendingInfo.Amount, sendingInfo.Receiver, nonce)
	if err != nil {
		return nil, err
	}

	err = e.Client.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, err
	}

	hashID := signedTx.Hash()

	return &hashID, nil
}

func (e *EthClient) SendERC20(sendingInfo SendingInfo, nonce uint64) (*common.Hash, error) {
	ctx := context.Background()

	fromAddress := wallet.GenWalletByKey(sendingInfo.PvKey)

	nonceTemp := big.NewInt(int64(nonce))

	gasPrice, err := e.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}
	// TODO 同步到上层gas fee 计算逻辑
	// gasPrice = 1.5 * gasPrice
	//gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(3))
	//gasPrice = new(big.Int).Quo(gasPrice, big.NewInt(2))
	chainId := e.ChainID

	// 获取auth
	auth, err := bind.NewKeyedTransactorWithChainID(sendingInfo.PvKey, chainId)
	if err != nil {
		return nil, err
	}

	tokenRaw := common.HexToAddress(sendingInfo.TokenAddr)

	tokenOBJ, err := build.NewToken(tokenRaw, e.Client)
	if err != nil {
		return nil, err
	}

	// 生成transfer()方法的hash值
	transferFnSignature := []byte("transfer(address,uint256)")
	hash := crypto.Keccak256(transferFnSignature)
	methodID := hash[:4]
	//fmt.Println(hexutil.Encode(methodID)) // 0xa9059cbb

	sendingInfo.Amount.SetString(sendingInfo.Amount.String(), 10) // 1000 tokens
	//fmt.Printf("Amount to string:%s\n", amount)
	paddedAmount := common.LeftPadBytes(sendingInfo.Amount.Bytes(), 32)
	//fmt.Println(hexutil.Encode(paddedAmount))

	paddedAddress := common.LeftPadBytes(sendingInfo.Receiver.Bytes(), 32)
	//fmt.Println(hexutil.Encode(paddedAddress))

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	gasLimit, err := e.Client.EstimateGas(ctx, ethereum.CallMsg{
		To:   &sendingInfo.Receiver,
		Data: data,
	})
	if err != nil {
		return nil, err
	}
	//fmt.Printf("Estimate gas limit:%d\n", gasLimit)

	if gasLimit <= cnt.GASLIMIT_ERC20 {
		gasLimit = cnt.GASLIMIT_ERC20
	}

	balance, err := tokenOBJ.BalanceOf(&bind.CallOpts{}, fromAddress)
	if err != nil {
		return nil, err
	}
	if sendingInfo.Amount.Cmp(balance) == 1 {
		return nil, xerrors.New("Transfer amount beyond account balance")
	}

	tx, err := tokenOBJ.Transfer(
		&bind.TransactOpts{
			From:     fromAddress,
			Nonce:    nonceTemp,
			Signer:   auth.Signer,
			Value:    big.NewInt(0),
			GasPrice: gasPrice,
			GasLimit: gasLimit,
			Context:  ctx,
			NoSend:   false,
		},
		sendingInfo.Receiver,
		sendingInfo.Amount,
	)
	if err != nil {
		return nil, err
	}

	hashID := tx.Hash()

	return &hashID, nil
}

//	重复多次请求发送交易请求 保证不会因为网络问题导致上链不成功,避免了跳nonce的问题
//
// SendingInfo TokenAddr 为空就是发送ETH, 有值就是发送ERC20
func (e *EthClient) SendTx(sendingInfo SendingInfo) (*common.Hash, error) {
	var err error
	for i := 0; i < 10; i++ {
		fromAddress := wallet.GenWalletByKey(sendingInfo.PvKey)
		nonce, errTemp := e.PendingNonce(fromAddress)
		if errTemp != nil {
			return nil, errTemp
		}
		//sleep 10s
		// fmt.Print("waiting for 10s")
		// time.Sleep(time.Second * 10)
		// 发送交易
		var hashID *common.Hash
		if sendingInfo.TokenAddr == "" {
			hashID, errTemp = e.SendETH(sendingInfo, nonce)
		} else {
			hashID, errTemp = e.SendERC20(sendingInfo, nonce)
		}
		// hashID, errTemp := e.SendERC20(sendingInfo, nonce)
		if errTemp == nil {
			return hashID, errTemp
		} else if reflect.TypeOf(errTemp).String() == "*rpc.jsonError" {
			err = errTemp
			if strings.Contains(errTemp.Error(), "nonce too low") {
				// 不退回队列情况,继续循环下一次，获取新的nonce
				// nonce 太低， 说明nonce 不能使用了，直接进入下次循环获取新的nonce
				err = errTemp
				continue
			} else {
				// 参数错误，或者其他原因，说明nonce还没被使用，退回队列,退出循环
				errRedis := redis.RejectNonce(nonce)
				if errRedis != nil {
					return nil, errRedis
				}
				return nil, err
			}
		} else if reflect.TypeOf(errTemp).String() == "*url.Error" {
			// 网络原因继续循环
			err = errTemp
			logger.Error("SendTx network error,try again,error:", errTemp)
			time.Sleep(time.Second * 2)
		} else {
			//其他未知原因，先记录，退出循环
			logger.Error("SendTx unkown error:", errTemp)
			return nil, errTemp
		}
	}
	return nil, err
}
func (e *EthClient) SendEthOffSign(amount *big.Int, receiver common.Address) (*common.Hash, uint64, error) {
	ctx := context.Background()

	signedTx, err := e.GenEthOffSign(ctx, amount, receiver)
	if err != nil {
		return nil, 0, err
	}

	height, err := e.Client.BlockNumber(ctx)
	if err != nil {
		return nil, 0, err
	}

	err = e.Client.SendTransaction(ctx, signedTx)
	if err != nil {
		logger.Errorf("SendEthOffSign Error:%v, \n receiver: %s \n tx info: %+v", err, receiver.Hex(), signedTx.Hash())
		return nil, 0, err
	}

	hashID := signedTx.Hash()

	return &hashID, height, nil
}

func (e *EthClient) SendERC20OffSign(amount *big.Int, receiver common.Address, tokenAddr string) (*common.Hash, uint64, error) {
	ctx := context.Background()

	//fromAddress := wallet.GenWalletByKey(pvKey)
	fromAddress := common.HexToAddress(EthCoreWalletAddr)

	n, err := e.PendingNonce(fromAddress)
	if err != nil {
		return nil, 0, err
	}
	nonce := big.NewInt(int64(n))

	gasPrice, err := e.SuggestGasPrice(ctx)
	if err != nil {
		return nil, 0, err
	}
	//gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(3))
	//gasPrice = new(big.Int).Quo(gasPrice, big.NewInt(2))

	height, err := e.Client.BlockNumber(ctx)
	if err != nil {
		return nil, 0, err
	}

	tokenRaw := common.HexToAddress(tokenAddr)

	tokenOBJ, err := build.NewToken(tokenRaw, e.Client)
	if err != nil {
		return nil, 0, err
	}

	// 生成transfer()方法的hash值
	transferFnSignature := []byte("transfer(address,uint256)")
	hash := crypto.Keccak256(transferFnSignature)
	methodID := hash[:4]
	//fmt.Println(hexutil.Encode(methodID)) // 0xa9059cbb

	amount.SetString(amount.String(), 10) // 1000 tokens
	//fmt.Printf("Amount to string:%s\n", amount)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	//fmt.Println(hexutil.Encode(paddedAmount))

	paddedAddress := common.LeftPadBytes(receiver.Bytes(), 32)
	//fmt.Println(hexutil.Encode(paddedAddress))

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	gasLimit, err := e.Client.EstimateGas(ctx, ethereum.CallMsg{
		To:   &receiver,
		Data: data,
	})
	if err != nil {
		return nil, 0, err
	}
	//fmt.Printf("Estimate gas limit:%d\n", gasLimit)

	if gasLimit <= cnt.GASLIMIT_ERC20 {
		gasLimit = cnt.GASLIMIT_ERC20
	}

	balance, err := tokenOBJ.BalanceOf(&bind.CallOpts{}, fromAddress)
	if err != nil {
		return nil, 0, err
	}
	if amount.Cmp(balance) == 1 {
		return nil, 0, xerrors.New("Transfer amount beyond account balance")
	}

	tx, err := tokenOBJ.Transfer(
		&bind.TransactOpts{
			From:     fromAddress,
			Nonce:    nonce,
			Signer:   e.getRemoteSignFn(),
			Value:    big.NewInt(0),
			GasPrice: gasPrice,
			GasLimit: gasLimit,
			Context:  ctx,
			NoSend:   false,
		},
		receiver,
		amount,
	)
	if err != nil {
		return nil, 0, err
	}

	hashID := tx.Hash()

	return &hashID, height, nil
}

func (e *EthClient) GenEthOffSign(ctx context.Context, amount *big.Int, receiver common.Address) (*types.Transaction, error) {
	fromAddress := common.HexToAddress(EthCoreWalletAddr)

	nonce, err := e.PendingNonce(fromAddress)
	if err != nil {
		return nil, err
	}

	gasPrice, err := e.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}
	// TODO 同步到上层gas fee 计算逻辑
	// gasPrice = 1.5 * gasPrice
	//gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(3))
	//gasPrice = new(big.Int).Quo(gasPrice, big.NewInt(2))

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &receiver,
		Value:    amount,
		Gas:      cnt.GASLIMIT_ETH, // default gas limit
		GasPrice: gasPrice,
		Data:     nil,
	})

	signedTx, err := e.getRemoteSignFn()(fromAddress, tx)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

// EncodeTX 编码离线签名Transaction
func (e *EthClient) EncodeTX(tx *types.Transaction, pvKey *ecdsa.PrivateKey) (string, error) {
	rawTxBytes, err := tx.MarshalJSON()
	if err != nil {
		return "", err
	}

	client, conn, err := signClient.NewClient()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	signResp, err := client.EthSign(context.Background(), &offlineSign.EthSignReq{
		TxBinaryText: string(rawTxBytes),
		ChainID:      e.ChainID.Int64(),
	})
	if err != nil {
		return "", err
	}

	return signResp.SignedTxBinary, nil
}

// DecodeTX 离线签名Transaction
func (e *EthClient) DecodeTX(rawTx string) (*types.Transaction, error) {
	tx := new(types.Transaction)

	err := tx.UnmarshalJSON([]byte(rawTx))
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (e *EthClient) GasCost(ctx context.Context, gasType string) (*big.Int, error) {
	gasPrice, err := e.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	var cost *big.Int
	switch gasType {
	case "ETH":
		cost = new(big.Int).Mul(gasPrice, big.NewInt(cnt.GASLIMIT_ETH))
		break
	case "ERC20":
		cost = new(big.Int).Mul(gasPrice, big.NewInt(cnt.GASLIMIT_ERC20))
		break
	default:
		return nil, xerrors.New("gasType error, not supported type: " + gasType)
	}

	return cost, nil
}

func (e *EthClient) QueryLatestTXByCnt(ctx context.Context, contractAddr string, fromBlock int64, toBlock int64) (*types.Log, error) {
	address := common.HexToAddress(contractAddr)

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(fromBlock),
		ToBlock:   big.NewInt(toBlock),
		Addresses: []common.Address{address},
	}

	logs, err := e.Client.FilterLogs(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(logs) == 0 {
		return nil, err
	}

	return &logs[0], nil
}

// SuggestGasPrice 重写了 ethclient.SuggestGasPrice 方法，增加了矿工 Tip Gas Price
func (e *EthClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	gasPrice, err := e.Client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}
	// 获取 矿工 Tip Gas Price
	gasTipPrice, err := e.Client.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, err
	}
	logger.Debugf("===========gasPrice: %s, gasTipPrice: %s", gasPrice.String(), gasTipPrice.String())

	//// TODO 待测试，提高上链成功率
	//gasTipPrice = new(big.Int).Mul(gasTipPrice, big.NewInt(2))
	gasPrice = gasPrice.Add(gasPrice, gasTipPrice)

	return gasPrice, nil
}

func (e *EthClient) SendEthTransaction(ctx context.Context, pvKey *ecdsa.PrivateKey, amount *big.Int, receiver common.Address, nonce uint64) (*types.Transaction, error) {

	gasPrice, err := e.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &receiver,
		Value:    amount,
		Gas:      cnt.GASLIMIT_ETH, // default gas limit
		GasPrice: gasPrice,
		Data:     nil,
	})

	chainID, err := e.Client.NetworkID(ctx)
	if err != nil {
		return nil, err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), pvKey)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

// 同步Transaction hash链上状态
// 注意：需要等待一个区块，才能返回结果
// tx状态：0->Fail | 1->Success
func (e *EthClient) SyncTxReceipt(ctx context.Context, hash *common.Hash) (*types.Receipt, error) {
	return e.Client.TransactionReceipt(ctx, *hash)
}

func (e *EthClient) SyncPendingTxReceipt(ctx context.Context, hash *common.Hash) (*types.Transaction, bool, error) {
	return e.Client.TransactionByHash(ctx, *hash)
}

func (e *EthClient) BalanceOfETH(addr string) (string, error) {
	userAddr := common.HexToAddress(addr)
	balance, err := e.Client.BalanceAt(context.Background(), userAddr, nil)
	if err != nil {
		return "", err
	}

	return balance.String(), nil
}

func (e *EthClient) BalanceOfERC20(addr string, tokenAddr string) (string, error) {
	userAddr := common.HexToAddress(addr)
	tokenRaw := common.HexToAddress(tokenAddr)

	tokenOBJ, err := build.NewToken(tokenRaw, e.Client)
	if err != nil {
		return "", err
	}

	balance, err := tokenOBJ.BalanceOf(&bind.CallOpts{}, userAddr)
	if err != nil {
		return "", err
	}

	return balance.String(), nil
}

func (e *EthClient) getRemoteSignFn() bind.SignerFn {
	f := func(ca common.Address, tx *types.Transaction) (*types.Transaction, error) {
		rawTxBytes, err := tx.MarshalJSON()
		if err != nil {
			return nil, err
		}

		client, conn, err := signClient.NewClient()
		if err != nil {
			return nil, err
		}
		defer conn.Close()

		signResp, err := client.EthSign(context.Background(), &offlineSign.EthSignReq{
			TxBinaryText: string(rawTxBytes),
			ChainID:      e.ChainID.Int64(),
		})
		if err != nil {
			return nil, err
		}
		tx = new(types.Transaction)

		err = tx.UnmarshalJSON([]byte(signResp.SignedTxBinary))
		if err != nil {
			return nil, err
		}
		return tx, nil
	}
	return f
}

// GetAddrTransfers 获取地址的交易记录
// addr: 地址 0x开头
// fromHeight: 开始高度
// coinType: 币种类型 ETH / USDT
// coinValue: 币种数量, 0.03 (ETH)
// return: hash, gasFee - 单位 eth , err
func (e *EthClient) GetAddrTransfers(addr string, fromHeight int64, coinType string, coinValue string) (hash string, gasFee *big.Float, err error) {
	type Tx struct {
		BlockNum        string      `json:"blockNum"`
		Hash            string      `json:"hash"`
		From            string      `json:"from"`
		To              string      `json:"to"`
		Value           json.Number `json:"value"`
		Erc721TokenID   interface{} `json:"erc721TokenId"`
		Erc1155Metadata interface{} `json:"erc1155Metadata"`
		TokenID         interface{} `json:"tokenId"`
		Asset           string      `json:"asset"`
		Category        string      `json:"category"`
	}

	type txs struct {
		Transfers []Tx `json:"transfers"`
	}

	param := struct {
		FromBlock string `json:"fromBlock"`
		ToAddress string `json:"toAddress"`
		// category: ["external", "internal", "erc20", "erc721", "erc1155"],
		Category []string `json:"category"`
	}{
		hexutil.EncodeUint64(uint64(fromHeight)),
		addr,
		[]string{"external", "internal"}, // external 别的地址转入，internal 自己转自己，erc20 代币
	}
	if coinType == constant.CoinUsdt {
		param.Category = []string{"erc20"}
	}
	res := txs{}
	// eth_accounts
	err = e.Client.Client().CallContext(context.Background(), &res, "alchemy_getAssetTransfers", param)
	if err != nil {
		return "", nil, err
	}
	if len(res.Transfers) == 0 {
		return "", nil, errors.New("no transfer")
	}

	//fmt.Printf("Res:%v\n", res)

	// 寻找对应的交易
	targetTx := Tx{}
	for _, tmp := range res.Transfers {
		v := tmp

		// 1. 判断是否是对应的币种
		if v.Asset != coinType {
			logger.Error("Invalid coin type")
			continue
		}
		// 2. 判断是否是对应的数量 coinValue "0.03", v.Value 0.03
		cv, _ := utils.StrToBigFloat(coinValue)
		vv, _ := utils.StrToBigFloat(string(v.Value))
		// 判断tx value 小于 order in amount
		if vv.Cmp(cv) != 0 {
			logger.Info("Tx value small than order amount")
			continue
		}
		//logger.Debug(v.Value)
		targetTx = v
	}
	if targetTx.Hash == "" {
		return "", nil, errors.New("no target transfer")
	}
	// 获取交易的gasFee
	tx, err := e.Client.TransactionReceipt(context.Background(), common.HexToHash(targetTx.Hash))
	if err != nil {
		return "", nil, err
	}
	//tx.GasUsed
	gasFeeWei := new(big.Int).Mul(big.NewInt(int64(tx.GasUsed)), tx.EffectiveGasPrice)
	gasFee = utils.WeiToEth(gasFeeWei)
	return targetTx.Hash, gasFee, nil
}

// GetTransferByTxSign 根据交易签名查询交易信息
func (e *EthClient) GetTransferByTxSign(ctx context.Context, sign string) (*EthTransfer, error) {
	tx, pending, err := e.Client.TransactionByHash(ctx, common.HexToHash(sign))
	if err != nil {
		return nil, err
	}
	if pending {
		logger.Info("Pending tx")
	} else {
		logger.Info("Success tx")
	}
	//构造eth消息结构
	ethTx := &EthTransfer{
		Hash:   tx.Hash().String(),
		To:     tx.To().String(),
		Amount: decimal.NewFromBigInt(tx.Value(), 0),
	}
	return ethTx, err
}

// GetTransactionByTxSign 根据交易签名查询交易信息
func (be *BaseEthClient) GetTransactionByTxSign(ctx context.Context, hash string) (*BaseTx, error) {
	var transaction *BaseTx
	err := be.Client.CallContext(context.Background(), &transaction, "eth_getTransactionByHash", hash)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

//根据
