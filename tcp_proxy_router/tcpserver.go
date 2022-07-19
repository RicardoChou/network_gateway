package tcp_proxy_router

import (
	"context"
	"fmt"
	"github.com/zhj/go_gateway/dao"
	"github.com/zhj/go_gateway/reverse_proxy"
	"github.com/zhj/go_gateway/tcp_proxy_middleware"
	"github.com/zhj/go_gateway/tcp_server"
	"log"
	"net"
)

var tcpServerList = []*tcp_server.TcpServer{}

type tcpHandler struct{}

func (t *tcpHandler) ServeTCP(c context.Context, src net.Conn) {
	src.Write([]byte("tcpHandler\n"))
}

func TCPServerRun() {
	// 获取tcp服务列表
	tcpServiceList := dao.ServiceManagerHandler.GetTcpServiceList()
	for _, serviceItem := range tcpServiceList {
		tmpItem := serviceItem
		go func(serviceDetail *dao.ServiceDetail) {
			// 获取TCP的地址（主要是端口号）
			addr := fmt.Sprintf(":%d", serviceDetail.TCPRule.Port)
			lb, err := dao.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
			if err != nil {
				log.Fatalf(" [INFO] GetTcpLoadBalancer %v err:%v\n", addr, err)
				return
			}
			//构建路由及设置中间件
			router := tcp_proxy_middleware.NewTcpSliceRouter()
			router.Group("/").Use(
				tcp_proxy_middleware.TCPFlowCountMiddleware(),
				tcp_proxy_middleware.TCPFlowLimitMiddleware(),
				tcp_proxy_middleware.TCPWhiteListMiddleware(),
				tcp_proxy_middleware.TCPBlackListMiddleware(),
			)

			//构建回调handler
			routerHandler := tcp_proxy_middleware.NewTcpSliceRouterHandler(
				func(c *tcp_proxy_middleware.TcpSliceRouterContext) tcp_server.TCPHandler {
					return reverse_proxy.NewTcpLoadBalanceReverseProxy(c, lb)
				}, router)
			baseCtx := context.WithValue(context.Background(), "service", serviceDetail)
			tcpServer := &tcp_server.TcpServer{
				Addr:    addr,
				Handler: routerHandler,
				BaseCtx: baseCtx,
			}
			tcpServerList = append(tcpServerList, tcpServer)
			log.Printf(" [INFO] tcp_proxy_run %v\n", addr)
			if err := tcpServer.ListenAndServe(); err != nil && err != tcp_server.ErrServerClosed {
				log.Fatalf(" [INFO] tcp_proxy_run %v err:%v\n", addr, err)
			}
		}(tmpItem)
	}
}

// TCPServerStop 遍历所有TCP服务器并关闭
func TCPServerStop() {
	for _, tcpServer := range tcpServerList {
		tcpServer.Close()
		log.Printf(" [INFO] tcp_proxy_stop %v stopped\n", tcpServer.Addr)
	}
}
