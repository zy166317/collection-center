package price

import (
	cnt "collection-center/contract/constant"
	"collection-center/internal/btc"
	"collection-center/internal/logger"
	"collection-center/library/redis"
	"fmt"
	"golang.org/x/xerrors"
	"time"
)

type BtcSat struct {
	Name             string `json:"name"`
	Height           int    `json:"height"`
	Hash             string `json:"hash"`
	Time             string `json:"time"`
	LatestUrl        string `json:"latest_url"`
	PreviousHash     string `json:"previous_hash"`
	PreviousUrl      string `json:"previous_url"`
	PeerCount        int    `json:"peer_count"`
	UnconfirmedCount int    `json:"unconfirmed_count"`
	HighFeePerKb     int    `json:"high_fee_per_kb"`
	MediumFeePerKb   int    `json:"medium_fee_per_kb"`
	LowFeePerKb      int    `json:"low_fee_per_kb"`
	LastForkHeight   int    `json:"last_fork_height"`
	LastForkHash     string `json:"last_fork_hash"`
}

func SyncSats() error {
	//sat := btc.GetLatestBlockStats()
	_, err := redis.InsertChainData(cnt.PPB, "6666")
	if err != nil {
		errNotice := fmt.Sprintf("Insert GAS price error:%s\n", err)
		logger.Error(errNotice)

		return xerrors.New(errNotice)
	}

	logger.Infof("Insert BTC SAT to redis successful")

	return nil
}

func testSAT() {
	sat, _ := btc.GetLatestBlockStats()
	fmt.Printf("Latest sat:%v\n", sat)
}

func SyncSat(period time.Duration) {
	for {
		SyncSats()

		time.Sleep(period)
	}
}
