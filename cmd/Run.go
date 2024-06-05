package cmd

import (
	config2 "collection-center/config"
	"collection-center/internal/logger"
	redis2 "collection-center/library/redis"
	"collection-center/service"
	"collection-center/service/block"
	"collection-center/service/price"
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
	// 同步流动性锁仓数据 - 仅在启动时执行一次 - 只能同步运行
	redis2.InitLockLiquid()
	// redis mq
	// 启动队列消费者
	go func() {
		err := queue.FirstQueueConsumer()
		if err != nil {
			logger.Error(err)
		}
		logger.Warn("启动第一队列消费者完成")

		err = queue.SecondQueueConsumer()
		if err != nil {
			logger.Error(err)
		}
		logger.Warn("启动第二队列消费者完成")

		err = queue.CoreToUserQueueConsumer()
		if err != nil {
			logger.Error(err)
		}
		logger.Warn("启动core转账队列消费者完成")

		err = queue.ThirdQueueConsumer()
		if err != nil {
			logger.Error(err)
		}
		logger.Warn("启动第三队列消费者完成")
	}()

	// 同步价格并写入redis
	go price.SyncPrice(15 * time.Second)
	// 同步区块高度并写入redis
	go block.SyncBlockHeight(15 * time.Second)
	// 将order表中的数据同步到redis - 仅在启动时执行一次
	go service.SyncOrderDataOnceToRedis()
	// 同步BTC SAT
	//go price.SyncSat(30 * time.Second)

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
	//<-redis.RedisMq.StopConsuming()
	<-redis2.FirstQueue.StopConsuming()
	<-redis2.SecondQueue.StopConsuming()
	<-redis2.ThirdQueue.StopConsuming()
	<-redis2.CoreToUserQueue.StopConsuming()
	// 所有开启的队列需要在此处进行等待关闭
	logger.Info("shutdown task consumer complete")

	logger.Warn("Shutdown Server ...")
	if err := srv.Shutdown(context.Background()); err != nil {
		logger.Fatal("Server Shutdown: ", err)
	}
	logger.Warn("Server exited")
}
