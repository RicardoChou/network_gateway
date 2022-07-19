package public

import (
	"sync"
	"time"
)

// 实现流量计数器单例模式

// LoadBalancerHandler 暴露出去的Handler
var FlowCounterHandler *FlowCounter

// FlowCounter 流量计数器结构体
type FlowCounter struct {
	RedisFlowCountMap   map[string]*RedisFlowCountService
	RedisFlowCountSlice []*RedisFlowCountService
	Locker              sync.RWMutex
}

func NewFlowCounter() *FlowCounter {
	return &FlowCounter{
		RedisFlowCountMap:   map[string]*RedisFlowCountService{},
		RedisFlowCountSlice: []*RedisFlowCountService{},
		Locker:              sync.RWMutex{},
	}
}

func init() {
	FlowCounterHandler = NewFlowCounter()
}

// GetCounter 获取计数器
func (counter *FlowCounter) GetCounter(serverName string) (*RedisFlowCountService, error) {
	// 查询匹配的计数器
	for _, item := range counter.RedisFlowCountSlice {
		if item.AppID == serverName {
			return item, nil
		}
	}
	// 匹配不到则新建一个计数器
	newCounter := NewRedisFlowCountService(serverName, 1*time.Second)
	counter.RedisFlowCountSlice = append(counter.RedisFlowCountSlice, newCounter)
	// 对Map操作最好用锁
	counter.Locker.Lock()
	defer counter.Locker.Unlock()
	counter.RedisFlowCountMap[serverName] = newCounter
	return newCounter, nil
}
