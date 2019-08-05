package timer

import (
	"sync/atomic"

	"github.com/BlockABC/wallet-btc-service/common/log"
)

const (
	STOP int32 = 0
	RUN  int32 = 1
)

type TimerTask struct {
	Task  func() error
	State int32
}

func NewTask(oneTask func() error) *TimerTask {
	return &TimerTask{
		Task:  oneTask,
		State: STOP,
	}
}

func (task *TimerTask) GetTask() func() {
	return func() {
		if bEnter := atomic.CompareAndSwapInt32(&task.State, STOP, RUN); bEnter {
			if err := task.Task(); nil != err {
				log.Log.Error("run timer task ", task.Task, err)
			}
			for {
				if atomic.CompareAndSwapInt32(&task.State, RUN, STOP) {
					break
				}
			}
		}
	}
}
