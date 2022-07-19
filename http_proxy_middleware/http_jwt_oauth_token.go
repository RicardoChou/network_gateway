package http_proxy_middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/zhj/go_gateway/dao"
	"github.com/zhj/go_gateway/middleware"
	"github.com/zhj/go_gateway/public"
	"strings"
)

// HTTPJwtOAuthTokenMiddleware JWT验证Token中间件
func HTTPJwtOAuthTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 取出服务的信息
		serverInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serverInterface.(*dao.ServiceDetail)
		// jwt验证Token的过程
		// 1.decode Jwt Token
		// 2.通过app_id从app_list中取出app_info
		// 3.将app_info放到gin的context中

		// 取出token
		token := strings.ReplaceAll(c.GetHeader("Authorization"), "Bearer ", "")
		//fmt.Println("token", token)
		appMatched := false
		// decode Jwt Token
		if token != "" {
			claims, err := public.JwtDecode(token)
			if err != nil {
				middleware.ResponseError(c, 2002, err)
				c.Abort()
				return
			}
			//fmt.Println("claims.Issuer", claims.Issuer)
			appList := dao.AppManagerHandler.GetAppList()
			// 通过app_id从app_list中取出app_info
			for _, appInfo := range appList {
				if appInfo.AppID == claims.Issuer {
					// 将app_info放到gin的context中
					c.Set("app", appInfo)
					appMatched = true
					break
				}
			}
		}
		// 服务开启验证了并且没有匹配到服务则退出
		if serviceDetail.AccessControl.OpenAuth == 1 && !appMatched {
			middleware.ResponseError(c, 2003, errors.New("not match valid app"))
			c.Abort()
			return
		}
		c.Next()
	}
}
