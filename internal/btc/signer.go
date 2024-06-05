package btc

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func SignRawStringTransaction_local(script string, inputs Inputs) (string, error) {
	b, err := hex.DecodeString(script)
	if err != nil {
		return "", nil
	}

	rawTx, err := btcutil.NewTxFromBytes(b)
	if err != nil {
		return "", err
	}

	tx := rawTx.MsgTx()

	return SignRawTxTransaction(tx, inputs)
}

func SignRawTxTransaction(tx *wire.MsgTx, inputs Inputs) (string, error) {

	errTmpl := "invalid input:[%d] : %s"

	hasSegWit := false
	hasZeroAmount := false

	prevOutFetcherMaps := make(map[wire.OutPoint]*wire.TxOut)
	prevOutFetcher := txscript.NewMultiPrevOutFetcher(prevOutFetcherMaps)

	signUseInputsMap := make(map[string]signInput)
	for idx, input := range inputs {
		addr, err := ParseWIF(input.WIF)
		if err != nil {
			return "", fmt.Errorf(errTmpl, idx, err)
		}

		if input.Amount == 0 {
			hasZeroAmount = true
		}

		if input.SegWit {
			if input.RedeemScript != "" {
				return "", fmt.Errorf(errTmpl, idx, "SegWit and RedeemScript cannot be set at the same time")
			}
			hasSegWit = true
		}

		if hasZeroAmount && hasSegWit {
			return "", fmt.Errorf("a segwit transaction was detected, but some inputs amount is zero")
		}

		signUseInputsMap[fmt.Sprintf("%s:%d", input.TxId, input.VOut)] = signInput{
			input: input,
			addr:  addr,
		}
	}

	if hasSegWit && len(signUseInputsMap) != len(tx.TxIn) {
		return "", fmt.Errorf("SegWit transaction was detected, but the number of UTXOs from the source transaction does not match the number of signed messages entered")
	}

	// calcTxSignHashes
	matched := 0
	var signHashes *txscript.TxSigHashes
	for idx, input := range tx.TxIn {
		txId := input.PreviousOutPoint.String()
		signUseInput, ok := signUseInputsMap[txId]
		if !ok {
			if hasSegWit {
				return "", fmt.Errorf("SegWit transaction was detected and %s private key is required", txId)
			}
			continue
		}

		matched++

		var pkScript []byte

		if signUseInput.input.SegWit {
			pkScript = signUseInput.addr.witnessPkScript
		} else {
			if signUseInput.input.RedeemScript != "" {
				decodeString, err := hex.DecodeString(signUseInput.input.RedeemScript)
				if err != nil {
					return "", fmt.Errorf(errTmpl, idx, err)
				}
				isMultiSignRedeemScript, _ := txscript.IsMultisigScript(decodeString)
				if !isMultiSignRedeemScript {
					return "", fmt.Errorf(errTmpl, idx, "invalid MultiSign redeem-script")
				}
				multiSignAddressPubKeyHash, err := btcutil.NewAddressScriptHash(decodeString, mnet)
				if err != nil {
					return "", fmt.Errorf(errTmpl, idx, err)
				}
				pkScript, _ = txscript.PayToAddrScript(multiSignAddressPubKeyHash)
				signUseInput.redeemScript = decodeString
			} else {
				pkScript = signUseInput.addr.p2pkhPkScript
			}
		}

		if pkScript == nil {
			return "", fmt.Errorf("calc pkScript failed for %s", txId)
		}

		amount, err := btcutil.NewAmount(signUseInput.input.Amount)
		if err != nil {
			return "", fmt.Errorf(errTmpl, idx, "invalid amount, "+err.Error())
		}

		signUseInputsMap[txId] = signUseInput
		prevOutFetcher.AddPrevOut(
			input.PreviousOutPoint,
			wire.NewTxOut(int64(amount), pkScript),
		)
	}

	if matched == 0 {
		return "", fmt.Errorf("no any input matched, the transaction was not signed")
	}

	if hasSegWit {
		signHashes = txscript.NewTxSigHashes(tx, prevOutFetcher)
	}

	for idx, input := range tx.TxIn {
		txId := input.PreviousOutPoint.String()
		signUseInput, ok := signUseInputsMap[txId]
		if !ok {
			continue
		}

		output := prevOutFetcher.FetchPrevOutput(input.PreviousOutPoint)

		errPrefix := fmt.Sprintf("invalid input[%d] %s : ", idx, txId)

		if !signUseInput.input.SegWit {

			pkScript := output.PkScript

			var redeemScript []byte
			if signUseInput.input.RedeemScript != "" {
				redeemScript = signUseInput.redeemScript
			}

			signScript, err := txscript.SignTxOutput(
				mnet,
				tx,
				idx,
				pkScript,
				txscript.SigHashAll,
				newLookupKeyFunc(signUseInput.addr.PrivateKey, mnet),
				newScriptDbFunc(redeemScript),
				tx.TxIn[idx].SignatureScript,
			)
			if err != nil {
				return "", fmt.Errorf("%s sign error : %s", errPrefix, err)
			}

			tx.TxIn[idx].SignatureScript = signScript
		} else {
			signHash, err := txscript.WitnessSignature(
				tx,
				signHashes,
				idx,
				output.Value,
				output.PkScript,
				txscript.SigHashAll,
				signUseInput.addr.PrivateKey,
				true,
			)
			if err != nil {
				return "", fmt.Errorf("%s sign error : %s", errPrefix, err)
			}
			tx.TxIn[idx].Witness = signHash
		}

	}

	return serializeTx(tx)
}
