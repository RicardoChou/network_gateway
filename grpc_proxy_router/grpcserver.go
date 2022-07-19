package grpc_proxy_router

import (
	"fmt"
	"github.com/e421083458/grpc-proxy/proxy"
	"github.com/zhj/go_gateway/dao"
	"github.com/zhj/go_gateway/grpc_proxy_middleware"
	"github.com/zhj/go_gateway/reverse_proxy"
	"google.golang.org/grpc"
	"log"
	"net"
)

var grpcServerList = []*warpGrpcServer{}

// warpGrpcServer 包装后的grpc服务器
type warpGrpcServer struct {
	Addr string
	*grpc.Server
}

func GrpcServerRun() {
	// 获取grpc服务列表
	grpcServiceList := dao.ServiceManagerHandler.GetGrpcServiceList()
	for _, serviceItem := range grpcServiceList {
		tmpItem := serviceItem
		go func(serviceDetail *dao.ServiceDetail) {
			// 获取gRPC的地址（主要是端口号）
			addr := fmt.Sprintf(":%d", serviceDetail.GRPCRule.Port)
			lb, err := dao.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
			if err != nil {
				log.Fatalf(" [INFO] GetTcpLoadBalancer %v err:%v\n", addr, err)
				return
			}
			lis, err := net.Listen("tcp", addr)
			if err != nil {
				log.Fatalf(" [INFO] GrpcListen %v err:%v\n", addr, err)
			}
			grpcHandler := reverse_proxy.NewGrpcLoadBalanceHandler(lb)
			s := grpc.NewServer(
				grpc.ChainStreamInterceptor(
					grpc_proxy_middleware.GrpcFlowCountMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcFlowLimitMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcJwtAuthTokenMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcJwtFlowCountMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcJwtFlowLimitMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcWhiteListMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcBlackListMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcHeaderTransferMiddleware(serviceDetail),
				),
				grpc.CustomCodec(proxy.Codec()),
				grpc.UnknownServiceHandler(grpcHandler),
			)
			grpcServerList = append(grpcServerList, &warpGrpcServer{
				Addr:   addr,
				Server: s,
			})

			log.Printf(" [INFO] grpc_proxy_run %v\n", addr)
			if err := s.Serve(lis); err != nil {
				log.Fatalf(" [INFO] grpc_proxy_run %v err:%v\n", addr, err)
			}
		}(tmpItem)
	}
}

// GrpcServerStop 遍历所有gRPC服务器并关闭
func GrpcServerStop() {
	for _, grpcServer := range grpcServerList {
		grpcServer.GracefulStop()
		log.Printf(" [INFO] grpc_proxy_stop %v stopped\n", grpcServer.Addr)
	}
}
