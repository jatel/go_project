package timer

import (
	"runtime"

	"github.com/BlockABC/wallet-btc-service/common/config"
	"github.com/BlockABC/wallet-btc-service/notify"
	"github.com/BlockABC/wallet-btc-service/omni"
	"github.com/robfig/cron"
)

func StartTimer() error {
	task := cron.New()

	// interval repair block
	if err := task.AddFunc(config.Cfg.BtcOpt.RepairBlock, IntervalRepairBlock); nil != err {
		return err
	}

	// interval repair transaction
	if err := task.AddFunc(config.Cfg.BtcOpt.RepairTransaction, IntervalRepairTransaction); nil != err {
		return err
	}

	// interval repair unconfirmed transaction
	if err := task.AddFunc(config.Cfg.BtcOpt.RepairUnconfirmedTrx, IntervalRepairUnconfirmedTransaction); nil != err {
		return err
	}

	// interval repair all block
	if err := task.AddFunc(config.Cfg.BtcOpt.RepairAll, IntervalRepairAll); nil != err {
		return err
	}

	// interval repair omni block
	if err := task.AddFunc(config.Cfg.BtcOpt.RepairOmniBlock, IntervalRepairOmniBlock); nil != err {
		return err
	}

	// interval repair omni transaction
	if err := task.AddFunc(config.Cfg.BtcOpt.RepairOmniTransaction, IntervalRepairOmniTransaction); nil != err {
		return err
	}

	// interval repair unconfirmed transaction
	if err := task.AddFunc(config.Cfg.BtcOpt.RepairUnconfirmedOmniTrx, IntervalRepairUnconfirmedOmniTransaction); nil != err {
		return err
	}

	// interval repair all omni block
	if err := task.AddFunc(config.Cfg.BtcOpt.RepairAll, IntervalRepairOmniAll); nil != err {
		return err
	}

	// interval modify goruntine number
	if err := task.AddFunc(config.Cfg.BtcOpt.ModifyGoroutine, IntervalModifyGoroutine); nil != err {
		return err
	}

	task.Start()

	select {}
}

func IntervalRepairBlock() {
	NewTask(notify.RepairBlocks).GetTask()()
}

func IntervalRepairTransaction() {
	NewTask(notify.RepairTransactions).GetTask()()
}

func IntervalRepairUnconfirmedTransaction() {
	NewTask(notify.RepairUnconfirmedTransaction).GetTask()()
}

func IntervalRepairAll() {
	NewTask(notify.RepairAll).GetTask()()
}

func IntervalRepairOmniBlock() {
	NewTask(omni.RepairFailOmniBlockHeight).GetTask()()
}

func IntervalRepairOmniTransaction() {
	NewTask(omni.RepairFailTransactionHash).GetTask()()
}

func IntervalRepairUnconfirmedOmniTransaction() {
	NewTask(omni.RepairUnconfirmedOmniTransaction).GetTask()()
}

func IntervalRepairOmniAll() {
	NewTask(omni.RepairAll).GetTask()()
}

func IntervalModifyGoroutine() {
	funcModify := func() error {
		num := runtime.NumGoroutine()
		if num > config.Cfg.BtcOpt.MaxGoroutine {
			config.TrxGoroutineRuntime = config.TrxGoroutineRuntime / 2
			if 0 == config.TrxGoroutineRuntime {
				config.TrxGoroutineRuntime = 1
			}
		} else {
			runtimeNum := config.TrxGoroutineRuntime * 2
			if runtimeNum < config.Cfg.BtcOpt.TrxGoroutineNum {
				config.TrxGoroutineRuntime = runtimeNum
			} else {
				config.TrxGoroutineRuntime = config.Cfg.BtcOpt.TrxGoroutineNum
			}
		}
		return nil
	}
	NewTask(funcModify).GetTask()()
}
