package dao

import (
	"fmt"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/zhj/go_gateway/dto"
	"github.com/zhj/go_gateway/public"
	"net/http/httptest"
	"strings"
	"sync"
)

// ServiceDetail 服务详情信息结构体
type ServiceDetail struct {
	Info          *ServiceInfo   `json:"info" description:"基本信息"`
	HTTPRule      *HttpRule      `json:"http_rule" description:"http_rule"`
	TCPRule       *TcpRule       `json:"tcp_rule" description:"tcp_rule"`
	GRPCRule      *GrpcRule      `json:"grpc_rule" description:"grpc_rule"`
	LoadBalance   *LoadBalance   `json:"load_balance" description:"load_balance"`
	AccessControl *AccessControl `json:"access_control" description:"access_control"`
}

// ServiceManager 对应服务信息管理的结构体
type ServiceManager struct {
	// 服务名称和服务详情的匹配
	ServiceMap   map[string]*ServiceDetail
	ServiceSlice []*ServiceDetail
	Locker       sync.RWMutex
	init         sync.Once
	err          error
}

// NewServiceManager 暴露出去的New方法
func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		ServiceMap:   map[string]*ServiceDetail{},
		ServiceSlice: []*ServiceDetail{},
		Locker:       sync.RWMutex{},
		init:         sync.Once{},
	}
}

// init 将ServiceManagerHandler进行初始化
func init() {
	ServiceManagerHandler = NewServiceManager()
}

// ServiceManagerHandler 将ServiceManager以Handler的形式暴露出去
var ServiceManagerHandler *ServiceManager

// GetTcpServiceList 获取TCP服务列表
func (s *ServiceManager) GetTcpServiceList() []*ServiceDetail {
	var list []*ServiceDetail
	for _, serverItem := range s.ServiceSlice {
		// 这边需要tmpItem这个临时变量，因为for range 的指针是固定的
		// 如果不用临时变量，则始终都是最后一个值
		tmpItem := serverItem
		if tmpItem.Info.LoadType == public.LoadTypeTCP {
			list = append(list, tmpItem)
		}
	}
	return list
}

// GetGrpcServiceList 获取Grpc服务列表
func (s *ServiceManager) GetGrpcServiceList() []*ServiceDetail {
	var list []*ServiceDetail
	for _, serverItem := range s.ServiceSlice {
		// 这边需要tmpItem这个临时变量，因为for range 的指针是固定的
		// 如果不用临时变量，则始终都是最后一个值
		tmpItem := serverItem
		if tmpItem.Info.LoadType == public.LoadTypeGRPC {
			list = append(list, tmpItem)
		}
	}
	return list
}

// HTTPAccessMode 匹配HTTP服务
func (s *ServiceManager) HTTPAccessMode(c *gin.Context) (*ServiceDetail, error) {
	//1、前缀匹配 /abc ==> serviceSlice.rule
	//2、域名匹配 www.test.com ==> serviceSlice.rule
	//host c.Request.Host
	//path c.Request.URL.Path

	// 域名匹配
	host := c.Request.Host
	host = host[0:strings.Index(host, ":")]
	// 前缀匹配
	path := c.Request.URL.Path
	fmt.Println("path", path)
	// 遍历所有服务，找出匹配到的服务
	for _, serviceItem := range s.ServiceSlice {
		if serviceItem.Info.LoadType != public.LoadTypeHTTP {
			continue
		}
		if serviceItem.HTTPRule.RuleType == public.HTTPRuleTypeDomain {
			if serviceItem.HTTPRule.Rule == host {
				return serviceItem, nil
			}
		}
		if serviceItem.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL {
			if strings.HasPrefix(path, serviceItem.HTTPRule.Rule) {
				return serviceItem, nil
			}
		}
	}
	return nil, errors.New("no matched service")
}

// LoadOnce 将服务加载到内存
func (s *ServiceManager) LoadOnce() error {
	// 因为ServiceManager内部有sync.Once{}，所以init.Do()只会执行一次
	s.init.Do(func() {
		// 获取服务列表输入结构体各项参数
		params := &dto.ServiceListInput{PageNo: 1, PageSize: 99999}
		// 模拟生成上下文context
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		// 从数据库中读取ServiceInfo
		// 获取数据库连接池
		tx, err := lib.GetGormPool("default")
		if err != nil {
			s.err = err
			return
		}
		// 从数据库中分页读取基本信息
		serviceInfo := &ServiceInfo{}
		// 获取服务列表
		list, _, err := serviceInfo.PageList(c, tx, params)
		if err != nil {
			s.err = err
			return
		}
		// 对map进行操作的时候加锁
		s.Locker.Lock()
		defer s.Locker.Unlock()
		for _, listItem := range list {
			// 获取服务详情信息
			// 这边需要tmpItem这个临时变量，因为for range 的指针是固定的
			// 如果不用临时变量，则始终都是最后一个值
			tmpItem := listItem
			serviceDetail, err := tmpItem.ServiceDetail(c, tx, &tmpItem)
			//fmt.Println("serviceDetail")
			//fmt.Println(public.Obj2Json(serviceDetail))
			if err != nil {
				return
			}
			// 设置服务名和服务信息的对应关系
			s.ServiceMap[listItem.ServiceName] = serviceDetail
			// 保存所有的服务信息
			s.ServiceSlice = append(s.ServiceSlice, serviceDetail)
		}

	})
	return s.err
}
