package block

import (
	cnt "collection-center/contract/constant"
	"collection-center/internal/btc"
	"collection-center/internal/logger"
	"collection-center/internal/rpc"
	"collection-center/library/redis"
	"context"
	"fmt"
	"golang.org/x/xerrors"
	"strconv"
	"time"
)

func EthBlockHeight() error {
	ethRpc, err := rpc.NewEthRpc()
	if err != nil {
		return err
	}
	defer ethRpc.Close()

	height, err := ethRpc.Client.BlockNumber(context.Background())
	if err != nil {
		return err
	}

	_, err = redis.InsertChainData(cnt.ETH_HEIGHT, strconv.FormatUint(height, 10))
	if err != nil {
		errNotice := fmt.Sprintf("Insert ETH block height error:%s\n", err)
		logger.Error(errNotice)

		return xerrors.New(errNotice)
	}

	logger.Infof("Insert ETH block height to redis successful")

	return nil
}

func BtcBlockHeight() error {
	blockCount, err := btc.Client.GetBlockCount()
	if err != nil {
		logger.Errorf("Sync btc block height error%v\n", err)
		return err
	}

	_, err = redis.InsertChainData(cnt.BTC_HEIGHT, strconv.FormatUint(uint64(blockCount), 10))
	if err != nil {
		errNotice := fmt.Sprintf("Insert ETH block height error:%s\n", err)
		logger.Error(errNotice)

		return xerrors.New(errNotice)
	}

	logger.Infof("Insert BTC block height to redis successful")

	return nil
}

func SyncBlockHeight(period time.Duration) {
	for {
		err := EthBlockHeight()
		if err != nil {
			logger.Error("Sync ETH block height error:", err)
		}
		err = BtcBlockHeight()
		if err != nil {
			logger.Error("Sync BTC block height error:", err)
		}

		time.Sleep(period)
	}
}
