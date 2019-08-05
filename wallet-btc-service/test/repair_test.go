package test

import (
	"fmt"
	"testing"

	"github.com/BlockABC/wallet-btc-service/common/config"
	"github.com/BlockABC/wallet-btc-service/database"
	"github.com/BlockABC/wallet-btc-service/notify"
)

func TestRepairBlocks(t *testing.T) {
	// initialize database
	dbOpt := config.Cfg.DbOpt
	err := database.Initialize(dbOpt.Address, dbOpt.User, dbOpt.Password, dbOpt.DbName, dbOpt.MaxOpenConn, dbOpt.MaxIdleConn, dbOpt.MaxWaitTimeout)
	if err != nil {
		panic(err)
	}

	notify.InitRepairManager()
	if err := notify.StoreFailBlockHash("000000000000000000011210e37d42c7daee6067b7c67fa4e1bc9e538290926d"); nil != err {
		fmt.Println(err)
	}

	if err := notify.RepairBlocks(); nil != err {
		fmt.Println(err)
	}

	select {}
}

func TestRepairUnconfirmedTransaction(t *testing.T) {
	// initialize database
	dbOpt := config.Cfg.DbOpt
	err := database.Initialize(dbOpt.Address, dbOpt.User, dbOpt.Password, dbOpt.DbName, dbOpt.MaxOpenConn, dbOpt.MaxIdleConn, dbOpt.MaxWaitTimeout)
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}

	//repair
	if err := notify.RepairUnconfirmedTransaction(); nil != err {
		fmt.Println(err)
	}

	select {}
}


func TestRepairAll(t *testing.T) {
	dbOpt := config.Cfg.DbOpt
	err := database.Initialize(dbOpt.Address, dbOpt.User, dbOpt.Password, dbOpt.DbName, dbOpt.MaxOpenConn, dbOpt.MaxIdleConn, dbOpt.MaxWaitTimeout)
	if err != nil {
		panic(err)
	}

	notify.InitRepairManager()
	if err := notify.StoreFailBlockHash("000000005c51de2031a895adc145ee2242e919a01c6d61fb222a54a54b4d3089"); nil != err {
		fmt.Println(err)
	}

	if err := notify.StoreFailBlockHash("00000000bc919cfb64f62de736d55cf79e3d535b474ace256b4fbb56073f64db"); nil != err {
		fmt.Println(err)
	}

	if err := notify.StoreFailBlockHash("0000000031975f17c5642a9c7e53ae4201a70a6ba6363036b55497b8754d1866"); nil != err {
		fmt.Println(err)
	}

	if err := notify.RepairBlocks(); nil != err {
		fmt.Println(err)
	}

	for {
		notify.RepairAll()
	}
}