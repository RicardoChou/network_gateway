package http_proxy_middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zhj/go_gateway/dao"
	"github.com/zhj/go_gateway/middleware"
	"github.com/zhj/go_gateway/public"
)

// HTTPAccessModeMiddleware 基于请求的信息来匹配接入的方式
func HTTPAccessModeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 进行服务的匹配
		service, err := dao.ServiceManagerHandler.HTTPAccessMode(c)
		if err != nil {
			middleware.ResponseError(c, 1001, err)
			c.Abort()
			return
		}
		fmt.Println("matched service", public.Obj2Json(service))
		// 设置上下文信息
		c.Set("service", service)
		c.Next()
	}
}
