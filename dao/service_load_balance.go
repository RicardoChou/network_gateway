package dao

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zhj/go_gateway/public"
	"github.com/zhj/go_gateway/reverse_proxy/load_balance"
	"gorm.io/gorm"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// LoadBalance 负载均衡信息结构体
type LoadBalance struct {
	ID            int64  `json:"id" gorm:"primary_key"`
	ServiceID     int64  `json:"service_id" gorm:"column:service_id" description:"服务id	"`
	CheckMethod   int    `json:"check_method" gorm:"column:check_method" description:"检查方法 tcpchk=检测端口是否握手成功	"`
	CheckTimeout  int    `json:"check_timeout" gorm:"column:check_timeout" description:"check超时时间	"`
	CheckInterval int    `json:"check_interval" gorm:"column:check_interval" description:"检查间隔, 单位s		"`
	RoundType     int    `json:"round_type" gorm:"column:round_type" description:"轮询方式 round/weight_round/random/ip_hash"`
	IpList        string `json:"ip_list" gorm:"column:ip_list" description:"ip列表"`
	WeightList    string `json:"weight_list" gorm:"column:weight_list" description:"权重列表"`
	ForbidList    string `json:"forbid_list" gorm:"column:forbid_list" description:"禁用ip列表"`

	UpstreamConnectTimeout int `json:"upstream_connect_timeout" gorm:"column:upstream_connect_timeout" description:"下游建立连接超时, 单位s"`
	UpstreamHeaderTimeout  int `json:"upstream_header_timeout" gorm:"column:upstream_header_timeout" description:"下游获取header超时, 单位s	"`
	UpstreamIdleTimeout    int `json:"upstream_idle_timeout" gorm:"column:upstream_idle_timeout" description:"下游链接最大空闲时间, 单位s	"`
	UpstreamMaxIdle        int `json:"upstream_max_idle" gorm:"column:upstream_max_idle" description:"下游最大空闲链接数"`
}

// TableName 对应数据库中的表名
func (t *LoadBalance) TableName() string {
	return "gateway_service_load_balance"
}

// Find  方法获得数据库中http规则的信息
func (t *LoadBalance) Find(c *gin.Context, tx *gorm.DB, search *LoadBalance) (*LoadBalance, error) {
	model := &LoadBalance{}
	// where方法支持结构体查询
	err := tx.WithContext(c).Where(search).Find(model).Error
	return model, err
}

// Save 方法将数据保存的数据库中
func (t *LoadBalance) Save(c *gin.Context, tx *gorm.DB) error {
	// save方法支持结构体保存
	if err := tx.WithContext(c).Save(t).Error; err != nil {
		return err
	}
	return nil
}

// GetIPListByModel 获取IP列表
func (t *LoadBalance) GetIPListByModel() []string {
	return strings.Split(t.IpList, ",")
}

// GetWeightListByModel 获取权重列表
func (t *LoadBalance) GetWeightListByModel() []string {
	return strings.Split(t.WeightList, ",")
}

// LoadBalancerHandler 暴露出去的Handler
var LoadBalancerHandler *LoadBalancer

// LoadBalancer 基于服务类型来建立负载均衡的方法
type LoadBalancer struct {
	LoadBalanceMap   map[string]*LoadBalancerItem
	LoadBalanceSlice []*LoadBalancerItem
	Locker           sync.RWMutex
}

type LoadBalancerItem struct {
	LoadBalance load_balance.LoadBalance
	ServiceName string
}

func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		LoadBalanceMap:   map[string]*LoadBalancerItem{},
		LoadBalanceSlice: []*LoadBalancerItem{},
		Locker:           sync.RWMutex{},
	}
}

func init() {
	LoadBalancerHandler = NewLoadBalancer()
}

// GetLoadBalancer 通过服务信息获取负载均衡器
func (l *LoadBalancer) GetLoadBalancer(service *ServiceDetail) (load_balance.LoadBalance, error) {
	// 匹配服务对应的负载均衡器，找到则直接返回
	for _, lbrItem := range l.LoadBalanceSlice {
		if lbrItem.ServiceName == service.Info.ServiceName {
			return lbrItem.LoadBalance, nil
		}
	}
	// 找不到服务匹配的负载均衡器则创建一个
	// 确定是http还是https
	schema := "http://"
	if service.HTTPRule.NeedHttps == 1 {
		schema = "https://"
	}
	if service.Info.LoadType==public.LoadTypeTCP || service.Info.LoadType==public.LoadTypeGRPC{
		schema = ""
	}
	// 获取IP列表和对应的权重的列表
	ipList := service.LoadBalance.GetIPListByModel()
	weightList := service.LoadBalance.GetWeightListByModel()
	// 组装IP列表和对应的权重
	ipConf := map[string]string{}
	for ipIndex, ipItem := range ipList {
		ipConf[ipItem] = weightList[ipIndex]
	}
	mConf, err := load_balance.NewLoadBalanceCheckConf(fmt.Sprintf("%s%s", schema, "%s"), ipConf)
	if err != nil {
		return nil, err
	}
	lb := load_balance.LoadBanlanceFactorWithConf(load_balance.LbType(service.LoadBalance.RoundType), mConf)

	//将新建的负载均衡器保存到Map和Slice中
	lbItem := &LoadBalancerItem{
		LoadBalance: lb,
		ServiceName: service.Info.ServiceName,
	}
	l.LoadBalanceSlice = append(l.LoadBalanceSlice, lbItem)

	// 对Map操作最好使用锁
	l.Locker.Lock()
	defer l.Locker.Unlock()
	l.LoadBalanceMap[service.Info.ServiceName] = lbItem
	return lb, nil
}

// TransporterHandler 暴露出去的Handler
var TransporterHandler *Transporter

// Transporter 连接池结构体
type Transporter struct {
	TransportMap   map[string]*TransportItem
	TransportSlice []*TransportItem
	Locker         sync.RWMutex
}

type TransportItem struct {
	Trans       *http.Transport
	ServiceName string
}

func NewTransporter() *Transporter {
	return &Transporter{
		TransportMap:   map[string]*TransportItem{},
		TransportSlice: []*TransportItem{},
		Locker:         sync.RWMutex{},
	}
}

func init() {
	TransporterHandler = NewTransporter()
}

// GetTrans 获取定制化连接池
func (t *Transporter) GetTrans(service *ServiceDetail) (*http.Transport, error) {
	// 匹配服务对应的连接池，找得到就直接返回连接池
	for _, transItem := range t.TransportSlice {
		if transItem.ServiceName == service.Info.ServiceName {
			return transItem.Trans, nil
		}
	}
	// 没找到连接池就新建一个定制的连接池并保存
	trans := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: time.Duration(service.LoadBalance.UpstreamConnectTimeout) * time.Second,
		}).DialContext,
		MaxIdleConns:          service.LoadBalance.UpstreamMaxIdle,
		IdleConnTimeout:       time.Duration(service.LoadBalance.UpstreamIdleTimeout) * time.Second,
		ResponseHeaderTimeout: time.Duration(service.LoadBalance.UpstreamHeaderTimeout) * time.Second,
	}

	//将新建的连接池保存到Map和Slice中
	transItem := &TransportItem{
		Trans:       trans,
		ServiceName: service.Info.ServiceName,
	}
	t.TransportSlice = append(t.TransportSlice, transItem)
	// 对Map操作最好使用锁
	t.Locker.Lock()
	defer t.Locker.Unlock()
	t.TransportMap[service.Info.ServiceName] = transItem
	return trans, nil
}
