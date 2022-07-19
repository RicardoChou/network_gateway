package dto

import (
	"github.com/gin-gonic/gin"
	"github.com/zhj/go_gateway/public"
	"time"
)

// AdminSessionInfo 管理员登陆Session信息结构体
type AdminSessionInfo struct {
	ID        int       `json:"id"`
	UserName  string    `json:"username"`
	LoginTime time.Time `json:"login_time"`
}

// AdminLoginInput 管理员登陆输入信息结构体
type AdminLoginInput struct {
	UserName string `json:"username" form:"username" comment:"管理员用户名" example:"admin" validate:"required,valid_username"` //管理员用户名
	Password string `json:"password" form:"password" comment:"密码" example:"123456" validate:"required"`                      //密码
}

// BindValidParam 验证参数有效性
func (param *AdminLoginInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, param)
}

// AdminLoginOutput 管理员登陆输出信息结构体
type AdminLoginOutput struct {
	Token string `json:"token" form:"token" comment:"token" example:"token" validate:""` //token
}
