package signClient

import (
	"collection-center/internal/signClient/pb/offlineSign"
	"context"
	"testing"
)

func TestSigner(t *testing.T) {
	SignerConfig = &RemoteSigner{
		Host:       "192.168.124.116",
		Port:       "8080",
		TlsPemPath: "../../resources/grpc-server-cert.pem",
		User:       "orca_off_signer",
		Pass:       "orca_6b9c1bb0f29f80a4fa759d8af2d26dd2",
	}

	client, conn, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	resp, err := client.EthSign(context.Background(), &offlineSign.EthSignReq{
		TxBinaryText: `{"type":"0x0","chainId":"0x5","nonce":"0x232","to":"0xa3c659e7384aa1baa4832ea7b616600661de22f3","gas":"0x5208","gasPrice":"0x15","maxPriorityFeePerGas":null,"maxFeePerGas":null,"value":"0x2386f26fc10000","input":"0x","v":"0x2e","r":"0x1d8b37d3ad06c665d688eb2819eb4b0d90ffef5100424d16fe7f6d7a8661f6ec","s":"0x2ae55736939cae6be35f059aedb4fce49f8425266dee33c9ed55f15a4b64bf81","hash":"0xc860ec7e1636667be9ad493ad9c484721375a0d9e4719d808a927663c30d6e10"}`,
		ChainID:      1,
	})
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(resp)
}
