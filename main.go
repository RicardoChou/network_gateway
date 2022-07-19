package main

import (
	"flag"
	"fmt"
	"github.com/e421083458/golang_common/lib"
	"github.com/zhj/go_gateway/dao"
	"github.com/zhj/go_gateway/grpc_proxy_router"
	"github.com/zhj/go_gateway/http_proxy_router"
	"github.com/zhj/go_gateway/router"
	"github.com/zhj/go_gateway/tcp_proxy_router"
	"os"
	"os/signal"
	"syscall"
)

// 使用同一个main文件来实现后台管理功能和代理服务器功能
// 通过分析flag参数来实现
// 终端 dashboard后台管理   server代理服务器
// 配置 ./conf/prod/ 对应配置文件夹

var (
	// 终端分两类：dashboard和server
	endpoint = flag.String("endpoint", "", "input endpoint dashboard or server")
	// 获取配置项(该行代码已集成在lib.InitModule源码中
	//config = flag.String("config", "", "input config file like ./conf/dev/")
	config = "./conf/dev/"
)

func main() {
	flag.Parse()
	if *endpoint == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *endpoint == "dashboard" {
		lib.InitModule(config, []string{"base", "mysql", "redis"})
		defer lib.Destroy()
		router.HttpServerRun()

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		router.HttpServerStop()
	} else {
		lib.InitModule(config, []string{"base", "mysql", "redis"})
		defer lib.Destroy()
		// 调用ServiceManagerHandler.LoadOnce()方法将服务加载到内存（只加载一次）
		dao.ServiceManagerHandler.LoadOnce()
		// 调用AppManagerHandler.LoadOnce()方法将租户加载到内存（只加载一次）
		dao.AppManagerHandler.LoadOnce()

		// 因为可能需要同时启动多个代理服务器，所以需要使用goroutine来启动
		go func() {
			http_proxy_router.HttpServerRun()
		}()
		go func() {
			http_proxy_router.HttpsServerRun()
		}()
		go func() {
			tcp_proxy_router.TCPServerRun()
		}()
		go func() {
			grpc_proxy_router.GrpcServerRun()
		}()
		fmt.Println("start server")
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		tcp_proxy_router.TCPServerStop()
		grpc_proxy_router.GrpcServerStop()
		http_proxy_router.HttpServerStop()
		http_proxy_router.HttpsServerStop()
	}
}
