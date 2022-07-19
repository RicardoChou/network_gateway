package dao

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AccessControl 服务信息权限控制结构体
type AccessControl struct {
	ID                int64  `json:"id" gorm:"primary_key"`
	ServiceID         int64  `json:"service_id" gorm:"column:service_id" description:"服务id"`
	OpenAuth          int    `json:"open_auth" gorm:"column:open_auth" description:"是否开启权限 1=开启"`
	BlackList         string `json:"black_list" gorm:"column:black_list" description:"黑名单ip"`
	WhiteList         string `json:"white_list" gorm:"column:white_list" description:"白名单ip"`
	WhiteHostName     string `json:"white_host_name" gorm:"column:white_host_name" description:"白名单主机"`
	ClientIPFlowLimit int    `json:"clientip_flow_limit" gorm:"column:clientip_flow_limit" description:"客户端ip限流"`
	ServiceFlowLimit  int    `json:"service_flow_limit" gorm:"column:service_flow_limit" description:"服务端限流"`
}

// TableName 对应数据库中的表名
func (t *AccessControl) TableName() string {
	return "gateway_service_access_control"
}

// Find  方法获得数据库中权限控制的信息
func (t *AccessControl) Find(c *gin.Context, tx *gorm.DB, search *AccessControl) (*AccessControl, error) {
	model := &AccessControl{}
	// where方法支持结构体查询
	err := tx.WithContext(c).Where(search).Find(model).Error
	return model, err
}

// Save 方法将数据保存的数据库中
func (t *AccessControl) Save(c *gin.Context, tx *gorm.DB) error {
	if err := tx.WithContext(c).Save(t).Error; err != nil {
		return err
	}
	return nil
}

// ListBYServiceID 方法将服务列表按照ID进行排序并返回
func (t *AccessControl) ListBYServiceID(c *gin.Context, tx *gorm.DB, serviceID int64) ([]AccessControl, int64, error) {
	// 服务列表
	var list []AccessControl
	// 服务总数
	var count int64
	query := tx.WithContext(c)
	// 先获取全部数据
	query = query.Table(t.TableName()).Select("*")
	// 筛选serviceID
	query = query.Where("service_id=?", serviceID)
	// 按照serviceID降序
	err := query.Order("id desc").Find(&list).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0, err
	}
	// 获取服务总数
	errCount := query.Count(&count).Error
	if errCount != nil {
		return nil, 0, err
	}
	return list, count, nil
}
