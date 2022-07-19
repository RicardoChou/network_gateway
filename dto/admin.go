package dto

import (
	"github.com/gin-gonic/gin"
	"github.com/zhj/go_gateway/public"
	"time"
)

// AdminInfoOutput 管理员信息输出结构体
type AdminInfoOutput struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	LoginTime    time.Time `json:"login_time"`
	Avatar       string    `json:"avatar"`
	Introduction string    `json:"introduction"`
	Roles        []string  `json:"roles"`
}

// ChangePwdInput 管理员密码修改输入结构体
type ChangePwdInput struct {
	Password string `json:"password" form:"password" comment:"密码" example:"123456" validate:"required"` //密码
}

// BindValidParam 验证参数有效性
func (param *ChangePwdInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, param)
}
