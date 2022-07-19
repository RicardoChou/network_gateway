package dao

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HttpRule http规则结构体
type HttpRule struct {
	ID             int64  `json:"id" gorm:"primary_key"`
	ServiceID      int64  `json:"service_id" gorm:"column:service_id" description:"服务id"`
	RuleType       int    `json:"rule_type" gorm:"column:rule_type" description:"匹配类型 domain=域名, url_prefix=url前缀"`
	Rule           string `json:"rule" gorm:"column:rule" description:"type=domain表示域名，type=url_prefix时表示url前缀"`
	NeedHttps      int    `json:"need_https" gorm:"column:need_https" description:"type=支持https 1=支持"`
	NeedWebsocket  int    `json:"need_websocket" gorm:"column:need_websocket" description:"启用websocket 1=启用"`
	NeedStripUri   int    `json:"need_strip_uri" gorm:"column:need_strip_uri" description:"启用strip_uri 1=启用"`
	UrlRewrite     string `json:"url_rewrite" gorm:"column:url_rewrite" description:"url重写功能，每行一个"`
	HeaderTransfor string `json:"header_transfor" gorm:"column:header_transfor" description:"header转换支持增加(add)、删除(del)、修改(edit) 格式: add headname headvalue"`
}

// TableName 对应数据库中的表名
func (t *HttpRule) TableName() string {
	return "gateway_service_http_rule"
}

// Find  方法获得数据库中http规则的信息
func (t *HttpRule) Find(c *gin.Context, tx *gorm.DB, search *HttpRule) (*HttpRule, error) {
	model := &HttpRule{}
	// where方法支持结构体查询
	err := tx.WithContext(c).Where(search).Find(model).Error
	return model, err
}

// FindFirst 方法获得数据库中第一个匹配的http规则的信息, 若不存在则返回ErrRecordNotFound
func (t *HttpRule) FindFirst(c *gin.Context, tx *gorm.DB, search *HttpRule) (*HttpRule, error) {
	model := &HttpRule{}
	// where方法支持结构体查询
	// todo 为什么tx.WithContext(c).Where(search).Find(model).Error查询失败返回的error是nil而
	// todo tx.WithContext(c).Where(search).First(model).Error查询失败返回的error是record not found？
	// 已经解决，Find方法返回nil，而First会返回ErrRecordNotFound
	err := tx.WithContext(c).Where(search).First(model).Error
	return model, err
}

// Save 方法将数据保存的数据库中
func (t *HttpRule) Save(c *gin.Context, tx *gorm.DB) error {
	// save方法支持结构体保存
	if err := tx.WithContext(c).Save(t).Error; err != nil {
		return err
	}
	return nil
}

// ListBYServiceID 方法将http服务规则列表按照ID进行排序并返回
func (t *HttpRule) ListBYServiceID(c *gin.Context, tx *gorm.DB, serviceID int64) ([]HttpRule, int64, error) {
	// 服务列表
	var list []HttpRule
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
