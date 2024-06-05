package crypto

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/xerrors"
	"strings"
)

func VerifySignature(signer string, signature string, msg string) (bool, error) {
	decodedSig, err := hexutil.Decode(signature)
	if err != nil {
		return false, err
	}

	if decodedSig[64] != 27 && decodedSig[64] != 28 {
		return false, xerrors.New("Invalid signature")
	}

	decodedSig[64] -= 27
	prefixedNonce := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(msg), msg)
	hash := crypto.Keccak256Hash([]byte(prefixedNonce))
	recoveredPublicKey, err := crypto.Ecrecover(hash.Bytes(), decodedSig)
	if err != nil {
		return false, err
	}

	secp256k1RecoveredPublicKey, err := crypto.UnmarshalPubkey(recoveredPublicKey)
	if err != nil {
		return false, err
	}

	recoveredAddress := crypto.PubkeyToAddress(*secp256k1RecoveredPublicKey).Hex()
	verifyStatus := strings.ToLower(signer) == strings.ToLower(recoveredAddress)

	return verifyStatus, nil
}
