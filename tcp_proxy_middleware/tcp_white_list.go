package tcp_proxy_middleware

import (
	"fmt"
	"github.com/zhj/go_gateway/dao"
	"github.com/zhj/go_gateway/public"
	"strings"
)

// TCPWhiteListMiddleware TCP白名单中间件
func TCPWhiteListMiddleware() func(c *TcpSliceRouterContext) {
	return func(c *TcpSliceRouterContext) {
		serverInterface := c.Get("service")
		if serverInterface == nil {
			c.conn.Write([]byte("get service empty"))
			c.Abort()
			return
		}
		serviceDetail := serverInterface.(*dao.ServiceDetail)
		splits := strings.Split(c.conn.RemoteAddr().String(), ":")
		clientIP := ""
		if len(splits) == 2 {
			clientIP = splits[0]
		}

		iplist := []string{}
		if serviceDetail.AccessControl.WhiteList != "" {
			iplist = strings.Split(serviceDetail.AccessControl.WhiteList, ",")
		}
		// 验证黑名单要前提是服务开启了验证
		if serviceDetail.AccessControl.OpenAuth == 1 && len(iplist) > 0 {
			if !public.InStringSlice(iplist, clientIP) {
				c.conn.Write([]byte(fmt.Sprintf("%s not in white ip list", clientIP)))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
