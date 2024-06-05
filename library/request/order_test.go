package request

import (
	"collection-center/config"
	"collection-center/library/utils"
	"fmt"
	"testing"
)

func TestVerifyEmail(t *testing.T) {
	err := utils.VerifyEmailFormat("sd@qq.com")
	if !err {
		t.Fatal("Invalid email address")
	}
	fmt.Print("Verify success\n")
}

func TestVerifyBtcWalletAddr(t *testing.T) {
	config.SetTestingConfig(&config.ServerConfig{
		Rpc: &config.Rpc{
			Test: true,
		},
	})
	err := VerifyBtcWalletAddr("2NFyQxSybFSQV55j8fXNBitczRKKTWvX2Ve")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Print("Verify success\n")
}

func TestVerifyEvmWalletAddr(t *testing.T) {
	//err := VerifyEvmWalletAddr("2NFyQxSybFSQV55j8fXNBitczRKKTWvX2Ve")
	err := VerifyEvmWalletAddr("0x22B6e3aBe7F2181D38e9c2F4d3Cd8F15ab1a1bfd")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Print("Verify success\n")
}

func TestVerifyNum(t *testing.T) {
	err := VerifyNum("0.123123")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Print("Verify success\n")
}
