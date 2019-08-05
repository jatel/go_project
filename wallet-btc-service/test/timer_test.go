package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/BlockABC/wallet-btc-service/timer"
	"github.com/robfig/cron"
)

func TestTimerTask(t *testing.T) {
	// one task function
	funcOne := func() error {
		fmt.Println(time.Now())
		time.Sleep(time.Duration(5) * time.Second)
		return nil
	}

	// timer task
	task := cron.New()

	if err := task.AddFunc("* * * * * *", timer.NewTask(funcOne).GetTask()); nil != err {
		fmt.Println(err)
	}

	task.Start()

	select {}
}
