package utils

import (
	"fmt"
	"testing"
)

func TestRandomRpcUrl(t *testing.T) {
	url, _, err := RandomEthRpcUrl([]string{""})
	if err != nil {
		t.Errorf("Get url error:%v", err)
	}

	fmt.Printf("Url:%s\n", url)
}
