package btc

import (
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
)

type Address struct {
	Address       string
	Bech32Address string
	PrivateKey    *btcec.PrivateKey
	PublicKey     *btcec.PublicKey
	AddressPubKey *btcutil.AddressPubKey
	WIF           string

	p2pkhPkScript   []byte
	witnessPkScript []byte
}

type MultiSigAddress struct {
	Asm       string   `json:"asm,omitempty,-"`
	Type      string   `json:"type,omitempty,-"`
	ReqSigs   int32    `json:"reqSigs,omitempty,-"`
	Address   string   `json:"address,omitempty,-"`
	Addresses []string `json:"addresses,omitempty,-"`
	Script    string   `json:"script,omitempty,-"`
}

type Transaction struct {
	Txid     string         `json:"txid"`
	Hash     string         `json:"hash"`
	Size     int64          `json:"size"`
	VSize    int64          `json:"vsize"`
	Weight   int64          `json:"weight"`
	Version  int32          `json:"version"`
	LockTime uint32         `json:"locktime"`
	Vins     []btcjson.Vin  `json:"vin"`
	Vouts    []btcjson.Vout `json:"vout"`
}

type Input struct {
	TxId         string
	VOut         int64
	WIF          string
	RedeemScript string
	SegWit       bool
	Amount       float64
}

type Output struct {
	PayToAddress string
	Amount       float64
}

type Inputs []Input
type Outputs []Output

type signInput struct {
	input        Input
	addr         *Address
	redeemScript []byte
}

type UTXO struct {
	TxID          string `json:"txid"`
	VOut          int64  `json:"vout"`
	Amount        string `json:"amount"`
	Satoshis      int64  `json:"satoshis"`
	Height        int64  `json:"height"`
	Confirmations int64  `json:"confirmations"`
}

type UTXOs []UTXO

type BODY struct {
	Errors []struct {
		Error string `json:"error"`
	} `json:"errors"`
	Tx struct {
		BlockHeight   int64     `json:"block_height"`
		BlockIndex    int64     `json:"block_index"`
		Hash          string    `json:"hash"`
		Addresses     []string  `json:"addresses"`
		Total         int64     `json:"total"`
		Fees          int64     `json:"fees"`
		Size          int64     `json:"size"`
		Vsize         int64     `json:"vsize"`
		Preference    string    `json:"preference"`
		RelayedBy     string    `json:"relayed_by"`
		Received      time.Time `json:"received"`
		Ver           int64     `json:"ver"`
		DoubleSpend   bool      `json:"double_spend"`
		VinSz         int64     `json:"vin_sz"`
		VoutSz        int64     `json:"vout_sz"`
		Confirmations int64     `json:"confirmations"`
		Inputs        []struct {
			PrevHash    string   `json:"prev_hash"`
			OutputIndex int64    `json:"output_index"`
			OutputValue int64    `json:"output_value"`
			Sequence    int64    `json:"sequence"`
			Addresses   []string `json:"addresses"`
			ScriptType  string   `json:"script_type"`
			Age         int64    `json:"age"`
		} `json:"inputs"`
		Outputs []struct {
			Value      int64    `json:"value"`
			Script     string   `json:"script"`
			Addresses  []string `json:"addresses"`
			ScriptType string   `json:"script_type"`
		} `json:"outputs"`
	} `json:"tx"`
	Tosign []string `json:"tosign"`
}

type BalanceStruct struct {
	Address            string `json:"address"`
	TotalReceived      int64  `json:"total_received"`
	TotalSent          int64  `json:"total_sent"`
	Balance            int64  `json:"balance"`
	UnconfirmedBalance int64  `json:"unconfirmed_balance"`
	FinalBalance       int64  `json:"final_balance"`
	NTx                int64  `json:"n_tx"`
	UnconfirmedNTx     int64  `json:"unconfirmed_n_tx"`
	FinalNTx           int64  `json:"final_n_tx"`
}

type TxDetails struct {
	BlockHash     string    `json:"block_hash"`
	BlockHeight   int64     `json:"block_height"`
	BlockIndex    int64     `json:"block_index"`
	Hash          string    `json:"hash"`
	Addresses     []string  `json:"addresses"`
	Total         int64     `json:"total"`
	Fees          int64     `json:"fees"`
	Size          int64     `json:"size"`
	Vsize         int64     `json:"vsize"`
	Preference    string    `json:"preference"`
	Confirmed     time.Time `json:"confirmed"`
	Received      time.Time `json:"received"`
	Ver           int64     `json:"ver"`
	DoubleSpend   bool      `json:"double_spend"`
	VinSz         int64     `json:"vin_sz"`
	VoutSz        int64     `json:"vout_sz"`
	DataProtocol  string    `json:"data_protocol"`
	Confirmations int64     `json:"confirmations"`
	Confidence    int64     `json:"confidence"`
}

type WalletDetails struct {
	Address            string   `json:"address"`
	TotalReceived      int64    `json:"total_received"`
	TotalSent          int64    `json:"total_sent"`
	Balance            int64    `json:"balance"`
	UnconfirmedBalance int64    `json:"unconfirmed_balance"`
	FinalBalance       int64    `json:"final_balance"`
	NTx                int64    `json:"n_tx"`
	UnconfirmedNTx     int64    `json:"unconfirmed_n_tx"`
	FinalNTx           int64    `json:"final_n_tx"`
	TxRefs             []TxRefs `json:"txrefs"`
}

type TxRefs struct {
	TxHash        string    `json:"tx_hash"`
	BlockHeight   int64     `json:"block_height"`
	TxInputN      int64     `json:"tx_input_n"`
	TxOutputN     int64     `json:"tx_output_n"`
	Value         int64     `json:"value"`
	RefBalance    int64     `json:"ref_balance"`
	Spent         bool      `json:"spent"`
	Confirmations int64     `json:"confirmations"`
	Confirmed     time.Time `json:"confirmed"`
	DoubleSpend   bool      `json:"double_spend"`
}
