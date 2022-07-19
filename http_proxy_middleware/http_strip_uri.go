package http_proxy_middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/zhj/go_gateway/dao"
	"github.com/zhj/go_gateway/middleware"
	"github.com/zhj/go_gateway/public"
	"strings"
)

// HTTPStripUriMiddleware 如果是前缀接入，则实现多余前缀的删除的中间件
func HTTPStripUriMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 取出服务的信息
		serviceInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		// 去除多余前缀如：
		//http://127.0.0.1:8080/test_http_string/abbb ->http://127.0.0.1:2004/abbb
		fmt.Println("c.Request.URL.Path", c.Request.URL.Path)
		if serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL && serviceDetail.HTTPRule.NeedStripUri == 1 {
			c.Request.URL.Path = strings.Replace(c.Request.URL.Path, serviceDetail.HTTPRule.Rule, "", 1)
		}
		fmt.Println("c.Request.URL.Path", c.Request.URL.Path)
		c.Next()
	}
}
