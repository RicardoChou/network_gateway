package dao

import (
	"github.com/gin-gonic/gin"
	"github.com/zhj/go_gateway/dto"
	"gorm.io/gorm"
	"time"
)

// ServiceInfo 服务信息
type ServiceInfo struct {
	ID          int64     `json:"id" gorm:"primary_key"`
	LoadType    int       `json:"load_type" gorm:"column:load_type" description:"负载类型 0=http 1=tcp 2=grpc"`
	ServiceName string    `json:"service_name" gorm:"column:service_name" description:"服务名称"`
	ServiceDesc string    `json:"service_desc" gorm:"column:service_desc" description:"服务描述"`
	UpdatedAt   time.Time `json:"create_at" gorm:"column:create_at" description:"更新时间"`
	CreatedAt   time.Time `json:"update_at" gorm:"column:update_at" description:"添加时间"`
	IsDelete    int8      `json:"is_delete" gorm:"column:is_delete" description:"是否已删除；0：否；1：是"`
}

// TableName 服务信息对应数据库中的表名
func (t *ServiceInfo) TableName() string {
	return "gateway_service_info"
}

// ServiceDetail 调用各种类型的服务的Find方法来获取其服务详情信息
func (t *ServiceInfo) ServiceDetail(c *gin.Context, tx *gorm.DB, search *ServiceInfo) (*ServiceDetail, error) {
	if search.ServiceName == "" {
		info, err := t.Find(c, tx, search)
		if err != nil {
			return nil, err
		}
		search = info
	}
	httpRule := &HttpRule{ServiceID: search.ID}
	httpRule, err := httpRule.Find(c, tx, httpRule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	tcpRule := &TcpRule{ServiceID: search.ID}
	tcpRule, err = tcpRule.Find(c, tx, tcpRule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	grpcRule := &GrpcRule{ServiceID: search.ID}
	grpcRule, err = grpcRule.Find(c, tx, grpcRule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	accessControl := &AccessControl{ServiceID: search.ID}
	accessControl, err = accessControl.Find(c, tx, accessControl)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	loadBalance := &LoadBalance{ServiceID: search.ID}
	loadBalance, err = loadBalance.Find(c, tx, loadBalance)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	detail := &ServiceDetail{
		Info:          search,
		HTTPRule:      httpRule,
		TCPRule:       tcpRule,
		GRPCRule:      grpcRule,
		LoadBalance:   loadBalance,
		AccessControl: accessControl,
	}
	return detail, nil
}

// GroupByLoadType 获取服务信息并按照类型分组
func (t *ServiceInfo) GroupByLoadType(c *gin.Context, tx *gorm.DB) ([]dto.DashboardServiceStatItemOutput, error) {
	// 首页大盘服务列表
	var list []dto.DashboardServiceStatItemOutput
	query := tx.WithContext(c)
	// 筛选出未被删除的服务
	query = query.Table(t.TableName()).Where("is_delete=0")
	// 按照loadType分组
	if err := query.Select("load_type, count(*) as value").Group("load_type").Scan(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// PageList 获取服务列表信息并分页
func (t *ServiceInfo) PageList(c *gin.Context, tx *gorm.DB, param *dto.ServiceListInput) ([]ServiceInfo, int64, error) {
	// 服务总数
	var total int64
	// 服务列表
	var list []ServiceInfo
	// 计算偏移量
	offset := (param.PageNo - 1) * param.PageSize
	query := tx.WithContext(c)
	// 排除已经删除的服务
	query = query.Table(t.TableName()).Where("is_delete=0")
	// 使用服务名称或者服务描述来进行模糊查询
	if param.Info != "" {
		query = query.Where("service_name like ? or service_desc like ?", "%"+param.Info+"%", "%"+param.Info+"%")
	}
	// 执行查询并按照id降序输出
	if err := query.Limit(param.PageSize).Offset(offset).
		Order("id desc").Find(&list).Error; err != nil &&
		err != gorm.ErrRecordNotFound {
		return nil, 0, err
	}
	// 获取总数
	query.Limit(param.PageSize).Offset(offset).Count(&total)
	return list, total, nil
}

// Find  方法获得数据库中服务信息
func (t *ServiceInfo) Find(c *gin.Context, tx *gorm.DB, search *ServiceInfo) (*ServiceInfo, error) {
	out := &ServiceInfo{}
	// where方法支持结构体查询
	err := tx.WithContext(c).Where(search).Find(out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FindFirst  方法获得数据库中第一个匹配的服务信息，若不存在则返回ErrRecordNotFound
func (t *ServiceInfo) FindFirst(c *gin.Context, tx *gorm.DB, search *ServiceInfo) (*ServiceInfo, error) {
	out := &ServiceInfo{}
	// where方法支持结构体查询
	err := tx.WithContext(c).Where(search).First(out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Save 方法将数据保存的数据库中
func (t *ServiceInfo) Save(c *gin.Context, tx *gorm.DB) error {
	return tx.WithContext(c).Save(t).Error
}
