package dao

import (
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/zhj/go_gateway/dto"
	"gorm.io/gorm"
	"net/http/httptest"
	"sync"
	"time"
)

// App 租户信息结构体
type App struct {
	ID        int64     `json:"id" gorm:"primary_key"`
	AppID     string    `json:"app_id" gorm:"column:app_id" description:"租户id"`
	Name      string    `json:"name" gorm:"column:name" description:"租户名称"`
	Secret    string    `json:"secret" gorm:"column:secret" description:"密钥"`
	WhiteIPS  string    `json:"white_ips" gorm:"column:white_ips" description:"ip白名单，支持前缀匹配"`
	Qpd       int64     `json:"qpd" gorm:"column:qpd" description:"日请求量限制"`
	Qps       int64     `json:"qps" gorm:"column:qps" description:"每秒请求量限制"`
	CreatedAt time.Time `json:"create_at" gorm:"column:create_at" description:"添加时间"`
	UpdatedAt time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	IsDelete  int8      `json:"is_delete" gorm:"column:is_delete" description:"是否已删除；0：否；1：是"`
}

// TableName 租户信息对应数据库中的表名
func (t *App) TableName() string {
	return "gateway_app"
}

// Find  方法获得数据库中租户信息
func (t *App) Find(c *gin.Context, tx *gorm.DB, search *App) (*App, error) {
	model := &App{}
	err := tx.WithContext(c).Where(search).Find(model).Error
	return model, err
}

// FindFirst  方法获得数据库中第一个匹配的租户信息，若不存在则返回ErrRecordNotFound
func (t *App) FindFirst(c *gin.Context, tx *gorm.DB, search *App) (*App, error) {
	model := &App{}
	err := tx.WithContext(c).Where(search).First(model).Error
	return model, err
}

// Save 方法将数据保存的数据库中
func (t *App) Save(c *gin.Context, tx *gorm.DB) error {
	if err := tx.WithContext(c).Save(t).Error; err != nil {
		return err
	}
	return nil
}

// AppList 获取租户列表信息并分页
func (t *App) AppList(c *gin.Context, tx *gorm.DB, param *dto.AppListInput) ([]App, int64, error) {
	// 服务总数
	var total int64
	// 服务列表
	var list []App
	// 计算偏移量
	offset := (param.PageNo - 1) * param.PageSize
	query := tx.WithContext(c)
	query = query.Table(t.TableName()).Select("*")
	// 排除已经删除的租户
	query = query.Where("is_delete = 0")
	// 使用租户名称或者租户ID来进行模糊查询
	if param.Info != "" {
		query = query.Where("name like ? or app_id like ?", "%"+param.Info+"%", "%"+param.Info+"%")
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

// AppManagerHandler 暴露出去的Handler
var AppManagerHandler *AppManager

// init  初始化AppManagerHandler
func init() {
	AppManagerHandler = NewAppManager()
}

// AppManager 对应租户信息管理的结构体
type AppManager struct {
	AppMap   map[string]*App
	AppSlice []*App
	Locker   sync.RWMutex
	init     sync.Once
	err      error
}

// NewAppManager 暴露出去的New方法
func NewAppManager() *AppManager {
	return &AppManager{
		AppMap:   map[string]*App{},
		AppSlice: []*App{},
		Locker:   sync.RWMutex{},
		init:     sync.Once{},
	}
}

// GetAppList 获取租户信息列表
func (s *AppManager) GetAppList() []*App {
	return s.AppSlice
}

// LoadOnce 将租户信息加载到内存
func (s *AppManager) LoadOnce() error {
	s.init.Do(func() {
		appInfo := &App{}
		// 模拟context的生成
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		tx, err := lib.GetGormPool("default")
		if err != nil {
			s.err = err
			return
		}
		params := &dto.AppListInput{PageNo: 1, PageSize: 99999}
		list, _, err := appInfo.AppList(c, tx, params)
		if err != nil {
			s.err = err
			return
		}
		s.Locker.Lock()
		defer s.Locker.Unlock()
		for _, listItem := range list {
			tmpItem := listItem
			s.AppMap[listItem.AppID] = &tmpItem
			s.AppSlice = append(s.AppSlice, &tmpItem)
		}
	})
	return s.err
}
