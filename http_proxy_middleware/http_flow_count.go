package http_proxy_middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/zhj/go_gateway/dao"
	"github.com/zhj/go_gateway/middleware"
	"github.com/zhj/go_gateway/public"
	"time"
)

// HTTPFlowCountMiddleware 流量计数器中间件
func HTTPFlowCountMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 取出服务的信息
		serviceInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		// 获取三种流量计数
		// 1、全站总流量统计
		// 2、当前服务的流量统计
		// 3、租户流量的统计（在HTTPJwtFlowCountMiddleware中实现）

		// 1、全站总流量统计
		totalCounter, err := public.FlowCounterHandler.GetCounter(public.FlowTotal)
		if err != nil {
			middleware.ResponseError(c, 4001, err)
			c.Abort()
			return
		}
		totalCounter.Increase()
		dayCount, _ := totalCounter.GetDayData(time.Now())
		fmt.Printf("totalCounter qps:%v,dayCount:%v", totalCounter.QPS, dayCount)

		// 2、当前服务的流量统计
		serviceCounter, err := public.FlowCounterHandler.GetCounter(public.FlowServicePrefix + serviceDetail.Info.ServiceName)
		if err != nil {
			middleware.ResponseError(c, 4002, err)
			c.Abort()
			return
		}
		serviceCounter.Increase()
		dayServiceCount, _ := serviceCounter.GetDayData(time.Now())
		fmt.Printf("serviceCounter qps:%v,dayCount:%v", serviceCounter.QPS, dayServiceCount)
		c.Next()
	}
}
