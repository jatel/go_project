package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/BlockABC/wallet-btc-service/httpserver"

	"github.com/BlockABC/wallet-btc-service/common/config"
	"github.com/BlockABC/wallet-btc-service/common/log"
	"github.com/BlockABC/wallet-btc-service/database"
)

func main() {
	log.Log.Notice("start bitcoin server database process...")
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
	// http server
	go httpserver.StartHttpServer(config.Cfg, ctx)

	wait()
}

//capture single
func wait() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(interrupt)
	sig := <-interrupt
	log.Log.Info("btc-wallet-server received signal:", sig)
}
