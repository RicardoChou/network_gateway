package http_proxy_middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/zhj/go_gateway/dao"
	"github.com/zhj/go_gateway/middleware"
	"regexp"
	"strings"
)

// HTTPUrlRewriteMiddleware URL重写功能中间件
func HTTPUrlRewriteMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 取出服务的信息
		serviceInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		for _, item := range strings.Split(serviceDetail.HTTPRule.UrlRewrite, ",") {
			fmt.Println("item rewrite", item)
			items := strings.Split(item, " ")
			if len(items) != 2 {
				continue
			}
			regexpItem, err := regexp.Compile(items[0])
			if err != nil {
				fmt.Println("regexp.Compile err", err)
				continue
			}
			fmt.Println("before rewrite", c.Request.URL.Path)
			replacePath := regexpItem.ReplaceAll([]byte(c.Request.URL.Path), []byte(items[1]))
			c.Request.URL.Path = string(replacePath)
			fmt.Println("after rewrite", c.Request.URL.Path)
		}
		c.Next()
	}
}
