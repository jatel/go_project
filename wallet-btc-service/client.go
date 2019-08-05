package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/BlockABC/wallet-btc-service/common/config"
	"github.com/BlockABC/wallet-btc-service/common/log"
	"github.com/BlockABC/wallet-btc-service/database"
	"github.com/BlockABC/wallet-btc-service/filestore"
	"github.com/BlockABC/wallet-btc-service/notify"
	"github.com/BlockABC/wallet-btc-service/omni"
	"github.com/BlockABC/wallet-btc-service/timer"
)

func main() {
	log.Log.Notice("start bitcoin client database process...")
	ctx := context.Background()

	// initialize database
	dbOpt := config.Cfg.DbOpt
	err := database.Initialize(dbOpt.Address, dbOpt.User, dbOpt.Password, dbOpt.DbName, dbOpt.MaxOpenConn, dbOpt.MaxIdleConn, dbOpt.MaxWaitTimeout)
	if err != nil {
		panic(err)
	}

	if err := database.InitRedis(config.Cfg.RedisOpt.RedisAddress, config.Cfg.RedisOpt.RedisDbNum); nil != err {
		panic(err)
	}

	// initialize repair
	notify.InitRepairManager()

	// set block repair begin
	if err, height := filestore.RepairStoreInstance.GetBlockBegin(); nil == err {
		notify.BlockHeightBegin = height
	}

	// set omni repair begin
	if err, height := filestore.RepairStoreInstance.GetOmniBegin(); nil == err {
		omni.OmniHeightBegin = height
	}

	// start timer task
	go timer.StartTimer()

	// start listen notification
	go notify.StartHttpServer(config.Cfg, ctx)

	wait()
}

//capture single
func wait() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(interrupt)
	sig := <-interrupt
	log.Log.Info("wallet-btc-client received signal:", sig)
}
