package sdk

import (
	"fmt"
	"github.com/Sheledons/config/config-sdk/pkg/log"
	"sync"
	"time"
)

/**
  @author: baisongyuan
  @since: 2022/7/28
**/

type Task func()

type TaskQueue struct {
	tickerMap map[string]*time.Ticker
	lock      sync.RWMutex
}

func newTaskQueue() *TaskQueue {
	return &TaskQueue{
		tickerMap: make(map[string]*time.Ticker),
	}
}
func (tq *TaskQueue) addTask(name string,period time.Duration, task Task) {
	tq.lock.Lock()
	defer tq.lock.Unlock()
	ticker := time.NewTicker(period)
	tq.tickerMap[name] = ticker
	go func() {
		for {
			<-ticker.C
			fmt.Println("执行定时任务")
			task()
		}
	}()
}
func (tq *TaskQueue) stop(name string)  {
	tq.lock.Lock()
	defer tq.lock.Unlock()
	if t, ok := tq.tickerMap[name]; ok {
		t.Stop()
	}
}
func (tq *TaskQueue) stopAll() {
	for name , t := range tq.tickerMap {
		t.Stop()
		log.Sugar.Infof("stop task: %s",name)
	}
}
