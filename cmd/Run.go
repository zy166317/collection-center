package cmd

import (
	config2 "collection-center/config"
	"collection-center/internal/logger"
	redis2 "collection-center/library/redis"
	"collection-center/service/queue"
	"collection-center/service/server"
	"context"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"
)

var RunCmd = &cli.Command{
	Name:  "run",
	Usage: "run --conf ./resources",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "conf", Usage: "--conf ./resources"},
	},
	Action: func(cctx *cli.Context) error {
		// Http server入口
		run(cctx)

		return nil
	},
}

func run(cctx *cli.Context) {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	//配置载入
	config2.InitConfig(cctx)
	//载入http server配置
	srv := server.NewHttpServer(config2.Config().Api.ListenPort, config2.Config().Api.Debug)
	logger.Infof("Server start at:" + strconv.Itoa(config2.Config().Api.ListenPort))
	//ethRpc, err := rpc.NewEthRpc()
	//if err != nil {
	//	logger.Error(err)
	//	return
	//}
	//ethRpc.GetTokenDecimal("0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599")

	// redis mq
	// 启动队列消费者
	go func() {
		err := queue.TonQueueConsumer()
		if err != nil {
			logger.Error(err)
			return
		}
	}()
	go func() {
		err := queue.EthQueueConsumer()
		if err != nil {
			logger.Error(err)
			return
		}
	}()
	go func() {
		err := queue.SolQueueConsumer()
		if err != nil {
			logger.Error(err)
			return
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGBUS, syscall.SIGKILL)
	done := make(chan byte, 1)
	go func() {
		for s := range quit {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				done <- '1'
				return
			default:

			}
		}
	}()
	<-done

	logger.Info("shutdown task consumer, please wait")
	<-redis2.SolQueue.StopConsuming()
	<-redis2.TonQueue.StopConsuming()
	<-redis2.ETHQueue.StopConsuming()
	// 所有开启的队列需要在此处进行等待关闭
	logger.Info("shutdown task consumer complete")

	logger.Warn("Shutdown Server ...")
	if err := srv.Shutdown(context.Background()); err != nil {
		logger.Fatal("Server Shutdown: ", err)
	}
	logger.Warn("Server exited")
}
