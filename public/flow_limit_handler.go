package public

import (
	"golang.org/x/time/rate"
	"sync"
)

// 实现限流器单例模式

// FlowLimiterHandler 暴露出去的Handler
var FlowLimiterHandler *FlowLimiter

// FlowLimiter 限流器结构体
type FlowLimiter struct {
	FlowLimiterMap   map[string]*FlowLimiterItem
	FlowLimiterSlice []*FlowLimiterItem
	Locker           sync.RWMutex
}

type FlowLimiterItem struct {
	ServiceName string
	Limiter     *rate.Limiter
}

func NewFlowLimiter() *FlowLimiter {
	return &FlowLimiter{
		FlowLimiterMap:   map[string]*FlowLimiterItem{},
		FlowLimiterSlice: []*FlowLimiterItem{},
		Locker:           sync.RWMutex{},
	}
}

func init() {
	FlowLimiterHandler = NewFlowLimiter()
}

// GetLimiter 获取限流器
func (limiter *FlowLimiter) GetLimiter(serverName string, qps float64) (*rate.Limiter, error) {
	// 查询匹配的限流器
	for _, item := range limiter.FlowLimiterSlice {
		if item.ServiceName == serverName {
			return item.Limiter, nil
		}
	}
	// 匹配不到则新建一个限流器
	newLimiter := rate.NewLimiter(rate.Limit(qps), int(qps*3))
	item := &FlowLimiterItem{
		serverName,
		newLimiter,
	}
	limiter.FlowLimiterSlice = append(limiter.FlowLimiterSlice, item)
	// 对Map操作最好用锁
	limiter.Locker.Lock()
	defer limiter.Locker.Unlock()
	limiter.FlowLimiterMap[serverName] = item
	return newLimiter, nil
}
