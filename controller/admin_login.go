package controller

import (
	"encoding/json"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/zhj/go_gateway/dao"
	"github.com/zhj/go_gateway/dto"
	"github.com/zhj/go_gateway/middleware"
	"github.com/zhj/go_gateway/public"
	"time"
)

type AdminLoginController struct{}

// AdminLoginRegister 注册管理员登陆和登出路由
func AdminLoginRegister(group *gin.RouterGroup) {
	adminLogin := &AdminLoginController{}
	group.POST("/login", adminLogin.AdminLogin)
	group.GET("/logout", adminLogin.AdminLogout)
}

// AdminLogin godoc
// @Summary 管理员登陆
// @Description 管理员登陆
// @Tags 管理员接口
// @ID /admin_login/login
// @Accept  json
// @Produce  json
// @Param body body dto.AdminLoginInput true "body"
// @Success 200 {object} middleware.Response{data=dto.AdminLoginOutput} "success"
// @Router /admin_login/login [post]
func (adminLogin *AdminLoginController) AdminLogin(c *gin.Context) {
	// 获取管理员登陆时输入的各项参数
	params := &dto.AdminLoginInput{}
	// 验证参数正确性
	if err := params.BindValidParam(c); err != nil {
		// 参数不正确返回错误信息并终止执行
		middleware.ResponseError(c, 2000, err)
		return
	}
	// 验证参数正确后执行
	// admin.LoginCheck内部逻辑：
	// 1. params.UserName 获取管理员的信息 adminInfo
	// 2. adminInfo.salt + params.Password sha256 -> saltPassword
	// 3. 验证saltPassword==adminInfo.Password
	// 获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	admin := &dao.Admin{}
	admin, err = admin.LoginCheck(c, tx, params)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	// 设置session
	sessInfo := &dto.AdminSessionInfo{
		ID:        admin.Id,
		UserName:  admin.UserName,
		LoginTime: time.Now(),
	}
	// json.Marshal将sessionInfo编码
	sessBts, err := json.Marshal(sessInfo)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	sess := sessions.Default(c)
	sess.Set(public.AdminSessionInfoKey, string(sessBts))
	// 保存session
	err = sess.Save()
	if err != nil {
		middleware.ResponseError(c, 2004, err)
		return
	}
	out := &dto.AdminLoginOutput{Token: admin.UserName}
	middleware.ResponseSuccess(c, out)
}

// AdminLogout godoc
// @Summary 管理员退出
// @Description 管理员退出
// @Tags 管理员接口
// @ID /admin_login/logout
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin_login/logout [get]
func (adminLogin *AdminLoginController) AdminLogout(c *gin.Context) {
	// 管理员登出功能只需要删除session即可
	sess := sessions.Default(c)
	// 删除session
	sess.Delete(public.AdminSessionInfoKey)
	// 保存删除session后的状态
	err := sess.Save()
	if err != nil {
		middleware.ResponseError(c, 2005, err)
		return
	}
	middleware.ResponseSuccess(c, "")
}
