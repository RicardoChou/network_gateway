package dao

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/zhj/go_gateway/dto"
	"github.com/zhj/go_gateway/public"
	"gorm.io/gorm"
	"time"
)

// Admin 管理员信息表
type Admin struct {
	Id        int       `json:"id" gorm:"primary_key" description:"自增主键"`
	UserName  string    `json:"user_name" gorm:"column:user_name" description:"管理员用户名"`
	Salt      string    `json:"salt" gorm:"column:salt" description:"盐"`
	Password  string    `json:"password" gorm:"column:password" description:"密码"`
	UpdatedAt time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	CreatedAt time.Time `json:"create_at" gorm:"column:create_at" description:"创建时间"`
	IsDelete  int       `json:"is_delete" gorm:"column:is_delete" description:"是否删除"`
}

// TableName 对应数据库中的表名
func (t *Admin) TableName() string {
	return "gateway_admin"
}

// LoginCheck 管理员用户信息验证，验证通过返回Admin信息
func (t *Admin) LoginCheck(c *gin.Context, tx *gorm.DB, param *dto.AdminLoginInput) (*Admin, error) {
	// 通过Admin.Find方法获得数据库中管理员的信息，第三个参数是自定义Admin结构体即可
	adminInfo, err := t.Find(c, tx, &Admin{UserName: param.UserName, IsDelete: 0})
	if err != nil {
		return nil, errors.New("用户信息不存在")
	}
	//param.Password
	//adminInfo.Salt
	// 通过adminInfo.Salt与param.Password生成加盐密码
	saltPassword := public.GenSaltPassword(adminInfo.Salt, param.Password)
	// 将生成的加盐密码与数据库中的密码比较
	if adminInfo.Password != saltPassword {
		return nil, errors.New("密码错误，请重新输入")
	}
	return adminInfo, nil
}

// Find  方法获得数据库中管理员的信息
func (t *Admin) Find(c *gin.Context, tx *gorm.DB, search *Admin) (*Admin, error) {
	out := &Admin{}
	// where方法支持结构体查询
	err := tx.WithContext(c).Where(search).Find(out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Save 方法将数据保存的数据库中
func (t *Admin) Save(c *gin.Context, tx *gorm.DB) error {
	// save方法支持结构体保存
	return tx.WithContext(c).Save(t).Error
}
