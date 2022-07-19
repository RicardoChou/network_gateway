package dao

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TcpRule tpc规则信息结构体
type TcpRule struct {
	ID        int64 `json:"id" gorm:"primary_key"`
	ServiceID int64 `json:"service_id" gorm:"column:service_id" description:"服务id"`
	Port      int   `json:"port" gorm:"column:port" description:"端口"`
}

// TableName 对应数据库中的表名
func (t *TcpRule) TableName() string {
	return "gateway_service_tcp_rule"
}

// Find  方法获得数据库中tcp规则的信息
func (t *TcpRule) Find(c *gin.Context, tx *gorm.DB, search *TcpRule) (*TcpRule, error) {
	model := &TcpRule{}
	// where方法支持结构体查询
	err := tx.WithContext(c).Where(search).Find(model).Error
	return model, err
}

// FindFirst  方法获得数据库中第一个匹配的tcp服务信息，若不存在则返回ErrRecordNotFound
func (t *TcpRule) FindFirst(c *gin.Context, tx *gorm.DB, search *TcpRule) (*TcpRule, error) {
	model := &TcpRule{}
	// where方法支持结构体查询
	err := tx.WithContext(c).Where(search).First(model).Error
	return model, err
}

// Save 方法将数据保存的数据库中
func (t *TcpRule) Save(c *gin.Context, tx *gorm.DB) error {
	// save方法支持结构体保存
	if err := tx.WithContext(c).Save(t).Error; err != nil {
		return err
	}
	return nil
}

// ListBYServiceID 方法将grpc服务规则列表按照ID进行排序并返回
func (t *TcpRule) ListBYServiceID(c *gin.Context, tx *gorm.DB, serviceID int64) ([]TcpRule, int64, error) {
	// 服务列表
	var list []TcpRule
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
