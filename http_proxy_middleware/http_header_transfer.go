package http_proxy_middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/zhj/go_gateway/dao"
	"github.com/zhj/go_gateway/middleware"
	"strings"
)

// HTTPHeaderTransMiddleware 实现请求头的重写
func HTTPHeaderTransMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 取出服务的信息
		serviceInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		for _, item := range strings.Split(serviceDetail.HTTPRule.HeaderTransfor, ",") {
			items := strings.Split(item, " ")
			// 请求头重写需要填写三个参数 add/edit headName headValue 或者del headName 任意数值
			if len(items) != 3 {
				continue
			}
			// 增加或者修改则直接覆盖请求头
			if items[0] == "add" || items[0] == "edit" {
				c.Request.Header.Set(items[1], items[2])
			}
			// 删除请求头
			if items[0] == "del" {
				c.Request.Header.Del(items[1])
			}
		}
		c.Next()
	}
}
