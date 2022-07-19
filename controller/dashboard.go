package controller

import (
	"fmt"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/zhj/go_gateway/dao"
	"github.com/zhj/go_gateway/dto"
	"github.com/zhj/go_gateway/middleware"
	"github.com/zhj/go_gateway/public"
	"time"
)

type DashboardController struct{}

// DashboardRegister 注册服务列表信息相关路由
func DashboardRegister(group *gin.RouterGroup) {
	dashboard := &DashboardController{}
	group.GET("/panel_group_data", dashboard.PanelGroupData)
	group.GET("/flow_stat", dashboard.FlowStat)
	group.GET("/service_stat", dashboard.ServiceStat)
}

// PanelGroupData godoc
// @Summary 指标统计
// @Description 指标统计
// @Tags 首页大盘
// @ID /dashboard/panel_group_data
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.PanelGroupDataOutput} "success"
// @Router /dashboard/panel_group_data [get]
func (Dashboard *DashboardController) PanelGroupData(c *gin.Context) {
	// 获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	// 从服务列表模块中取出服务总数
	serviceInfo := &dao.ServiceInfo{}
	_, serviceCount, err := serviceInfo.PageList(c, tx, &dto.ServiceListInput{PageSize: 1, PageNo: 1})
	fmt.Println("serviceCount:", serviceCount)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	// 从租户模块中取出租户总数
	appInfo := &dao.App{}
	_, appCount, err := appInfo.AppList(c, tx, &dto.AppListInput{PageNo: 1, PageSize: 1})
	fmt.Println("appCount:", appCount)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	// 获取服务计数器
	serviceCounter, err := public.FlowCounterHandler.GetCounter(public.FlowTotal)
	if err != nil {
		middleware.ResponseError(c, 2004, err)
		return
	}

	fmt.Println("CurrentQPS:", serviceCounter.QPS)
	fmt.Println("TodayRequestCount:", serviceCounter.TotalCount)
	// 组装返回信息
	out := &dto.PanelGroupDataOutput{
		ServiceCount:      serviceCount,
		AppCount:          appCount,
		CurrentQPS:        serviceCounter.QPS,
		TodayRequestCount: serviceCounter.TotalCount,
	}
	middleware.ResponseSuccess(c, out)
}

// FlowStat godoc
// @Summary 流量统计
// @Description 流量统计
// @Tags 首页大盘
// @ID /dashboard/flow_stat
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.FlowStatOutput} "success"
// @Router /dashboard/flow_stat [get]
func (Dashboard *DashboardController) FlowStat(c *gin.Context) {
	counter, err := public.FlowCounterHandler.GetCounter(public.FlowTotal)
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	// 昨天的服务统计
	yesterdayList := []int64{}
	currentTime := time.Now()
	yesterTime := currentTime.Add(-1 * time.Duration(time.Hour*24))
	for i := 0; i <= 23; i++ {
		dateTime := time.Date(yesterTime.Year(), yesterTime.Month(), yesterTime.Day(), i, 0, 0, 0, lib.TimeLocation)
		hourData, _ := counter.GetHourData(dateTime)
		yesterdayList = append(yesterdayList, hourData)
	}
	// 今天的服务统计
	var todayList []int64
	for i := 0; i <= currentTime.Hour(); i++ {
		dateTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), i, 0, 0, 0, lib.TimeLocation)
		hourData, _ := counter.GetHourData(dateTime)
		todayList = append(todayList, hourData)
	}

	middleware.ResponseSuccess(c, dto.ServiceStatOutput{
		Today:     todayList,
		Yesterday: yesterdayList,
	})
}

// ServiceStat godoc
// @Summary 服务统计
// @Description 服务统计
// @Tags 首页大盘
// @ID /dashboard/service_stat
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.DashboardServiceStatOutput} "success"
// @Router /dashboard/service_stat [get]
func (Dashboard *DashboardController) ServiceStat(c *gin.Context) {
	// 获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	// 从服务模块中取出按照类型分组后的服务信息
	serviceInfo := &dao.ServiceInfo{}
	list, err := serviceInfo.GroupByLoadType(c, tx)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	// 获取服务信息中的各类服务名称
	var legend []string
	for index, item := range list {
		name, ok := public.LoadTypeMap[item.LoadType]
		if !ok {
			middleware.ResponseError(c, 2003, errors.New("load_type not found"))
			return
		}
		list[index].Name = name
		legend = append(legend, name)
	}
	// 组装返回信息
	out := &dto.DashboardServiceStatOutput{
		Legend: legend,
		Data:   list,
	}
	middleware.ResponseSuccess(c, out)
}
