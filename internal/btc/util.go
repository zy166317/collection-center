package btc

import (
	"bytes"
	"collection-center/library/constant"
	"encoding/hex"
	"fmt"

	"math/rand"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func serializeTx(tx *wire.MsgTx) (string, error) {

	var txBuffer bytes.Buffer
	err := tx.Serialize(&txBuffer)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(txBuffer.Bytes()), nil
}
func CreateRawStringTransaction(inputs Inputs, outputs Outputs) (string, error) {
	transaction, err := CreateRawTransaction(inputs, outputs)
	if err != nil {
		return "", err
	}

	return serializeTx(transaction)
}
func CreateRawTransaction(inputs Inputs, outputs Outputs) (*wire.MsgTx, error) {
	// wire.TxVersion
	tx := wire.NewMsgTx(2)

	for _, input := range inputs {
		txHash, err := chainhash.NewHashFromStr(input.TxId)
		if err != nil {
			return nil, err
		}

		prevOut := wire.NewOutPoint(txHash, uint32(input.VOut))
		txIn := wire.NewTxIn(prevOut, nil, nil)
		tx.AddTxIn(txIn)
	}

	for _, output := range outputs {

		sendAmount, err := btcutil.NewAmount(output.Amount)
		if err != nil {
			return nil, err
		}
		address, err := btcutil.DecodeAddress(output.PayToAddress, mnet)
		if err != nil {
			return nil, err
		}

		pkScript, err := txscript.PayToAddrScript(address)
		if err != nil {
			return nil, err
		}

		out := wire.NewTxOut(int64(sendAmount), pkScript)
		tx.AddTxOut(out)
	}

	return tx, nil

}
func ParseWIF(key string) (*Address, error) {
	wif, err := btcutil.DecodeWIF(key)
	if err != nil {
		return nil, err
	}

	compressedPubKey := wif.PrivKey.PubKey().SerializeCompressed()
	addressPubKey, err := btcutil.NewAddressPubKey(compressedPubKey, mnet)
	if err != nil {
		return nil, err
	}

	// get p2pkhPkScript
	addressPubKeyHash, _ := btcutil.NewAddressPubKeyHash(btcutil.Hash160(compressedPubKey), mnet)
	p2pkhPkScript, _ := txscript.PayToAddrScript(addressPubKeyHash)
	// get witnessPkScript
	addressWitnessPubKeyHash, _ := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(compressedPubKey), mnet)
	witnessPkScript, _ := txscript.PayToAddrScript(addressWitnessPubKeyHash)

	return &Address{
		Address:       addressPubKey.EncodeAddress(),
		Bech32Address: addressWitnessPubKeyHash.EncodeAddress(),
		PrivateKey:    wif.PrivKey,
		PublicKey:     wif.PrivKey.PubKey(),
		WIF:           wif.String(),

		AddressPubKey: addressPubKey,

		p2pkhPkScript:   p2pkhPkScript,
		witnessPkScript: witnessPkScript,
	}, nil
}
func newLookupKeyFunc(privateKey *btcec.PrivateKey, params *chaincfg.Params) txscript.KeyDB {
	return txscript.KeyClosure(func(addr btcutil.Address) (*btcec.PrivateKey, bool, error) {
		// For validate MultiSig script
		privateKeyAddr, err := btcutil.NewAddressPubKey(privateKey.PubKey().SerializeCompressed(), params)
		if err != nil {
			return nil, false, err
		}
		if privateKeyAddr.EncodeAddress() != addr.EncodeAddress() {
			return nil, false, fmt.Errorf("private key not match")
		}
		// The normal transaction can be directly returned to the private key.
		return privateKey, true, nil
	})
}

// newScriptDbFunc only work for MultiSig
func newScriptDbFunc(pkScript []byte) txscript.ScriptDB {
	return txscript.ScriptClosure(func(address btcutil.Address) ([]byte, error) {
		return pkScript, nil
	})
}

func NewPrivateKey() *btcec.PrivateKey {
	key, _ := btcec.NewPrivateKey()

	return key
}

// 随机获取一个BCY的token
func GetRandomBCYToken() string {
	return constant.BCY_TOKENS[rand.Intn(len(constant.BCY_TOKENS))]
}
