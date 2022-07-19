package controller

import (
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/zhj/go_gateway/dao"
	"github.com/zhj/go_gateway/dto"
	"github.com/zhj/go_gateway/middleware"
	"github.com/zhj/go_gateway/public"
	"time"
)

type AppController struct{}

// AppRegister 租户路由注册
func AppRegister(router *gin.RouterGroup) {
	app := AppController{}
	router.GET("/app_list", app.AppList)
	router.GET("/app_detail", app.AppDetail)
	router.GET("/app_delete", app.AppDelete)
	router.GET("/app_stat", app.AppStatistics)
	router.POST("/app_add", app.AppAdd)
	router.POST("/app_update", app.AppUpdate)
}

// AppList godoc
// @Summary 租户列表
// @Description 租户列表
// @Tags 租户管理
// @ID /app/app_list
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param page_size query string true "每页多少条"
// @Param page_no query string true "页码"
// @Success 200 {object} middleware.Response{data=dto.AppListOutput} "success"
// @Router /app/app_list [get]
func (app *AppController) AppList(c *gin.Context) {
	// 获取租户列表输入结构体各项参数
	params := &dto.AppListInput{}
	// 验证参数正确性
	if err := params.BindValidParam(c); err != nil {
		// 参数不正确返回错误信息并终止执行
		middleware.ResponseError(c, 2000, err)
		return
	}
	// 从数据库中读取AppInfo
	// 获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	// 获取租户信息
	appInfo := &dao.App{}
	// 获取租户列表和租户总数
	list, total, err := appInfo.AppList(c, tx, params)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	// 根据返回的租户列表信息来组装返回值
	var outPutList []dto.AppListItemOutput
	for _, item := range list {
		appCounter, err := public.FlowCounterHandler.GetCounter(public.FlowAppPrefix + item.AppID)
		if err != nil {
			middleware.ResponseError(c, 2003, err)
			c.Abort()
			return
		}
		outPutList = append(outPutList, dto.AppListItemOutput{
			ID:       item.ID,
			AppID:    item.AppID,
			Name:     item.Name,
			Secret:   item.Secret,
			WhiteIPS: item.WhiteIPS,
			Qpd:      item.Qpd,
			Qps:      item.Qps,
			RealQpd:  appCounter.TotalCount,
			RealQps:  appCounter.QPS,
		})
	}
	output := dto.AppListOutput{
		List:  outPutList,
		Total: total,
	}
	middleware.ResponseSuccess(c, output)
	return
}

// AppDetail godoc
// @Summary 租户详情
// @Description 租户详情
// @Tags 租户管理
// @ID /app/app_detail
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} middleware.Response{data=dao.App} "success"
// @Router /app/app_detail [get]
func (app *AppController) AppDetail(c *gin.Context) {
	// 获取租户具体输入结构体各项参数
	params := &dto.AppDetailInput{}
	// 验证参数正确性
	if err := params.BindValidParam(c); err != nil {
		// 参数不正确返回错误信息并终止执行
		middleware.ResponseError(c, 2000, err)
		return
	}
	// todo 这里会使得已经被软删除的租户也可以查询出详情
	// todo 需要进行逻辑上的修正

	// 获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	// 通过租户ID来进行查询
	search := &dao.App{
		ID: params.ID,
	}

	// 获取租户详情信息
	appDetail, err := search.FindFirst(c, tx, search)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	middleware.ResponseSuccess(c, appDetail)
	return
}

// AppDelete godoc
// @Summary 租户删除
// @Description 租户删除
// @Tags 租户管理
// @ID /app/app_delete
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /app/app_delete [get]
func (app *AppController) AppDelete(c *gin.Context) {
	// 获取租户输入结构体各项参数
	params := &dto.AppDeleteInput{}
	// 验证参数正确性
	if err := params.BindValidParam(c); err != nil {
		// 参数不正确返回错误信息并终止执行
		middleware.ResponseError(c, 2000, err)
		return
	}

	// 获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	// 通过租户ID来进行查询
	search := &dao.App{
		ID: params.ID,
	}

	// 获取租户详情信息
	appInfo, err := search.FindFirst(c, tx, search)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	// 将服务信息中的IsDelete置为1，进行软删除
	appInfo.IsDelete = 1
	// 保存到数据库
	if err = appInfo.Save(c, tx); err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, "")
	return
}

// AppStatistics godoc
// @Summary 租户统计
// @Description 租户统计
// @Tags 租户管理
// @ID /app/app_stat
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} middleware.Response{data=dto.StatisticsOutput} "success"
// @Router /app/app_stat [get]
func (app *AppController) AppStatistics(c *gin.Context) {
	// 获取租户具体输入结构体各项参数
	params := &dto.AppDetailInput{}
	// 验证参数正确性
	if err := params.BindValidParam(c); err != nil {
		// 参数不正确返回错误信息并终止执行
		middleware.ResponseError(c, 2000, err)
		return
	}

	// 获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	// 通过租户ID来进行查询
	search := &dao.App{
		ID: params.ID,
	}

	// 获取租户详情信息
	appDetail, err := search.FindFirst(c, tx, search)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	//今日流量全天小时级访问统计
	todayStat := []int64{}
	counter, err := public.FlowCounterHandler.GetCounter(public.FlowAppPrefix + appDetail.AppID)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		c.Abort()
		return
	}
	currentTime := time.Now()
	for i := 0; i <= time.Now().In(lib.TimeLocation).Hour(); i++ {
		dateTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), i, 0, 0, 0, lib.TimeLocation)
		hourData, _ := counter.GetHourData(dateTime)
		todayStat = append(todayStat, hourData)
	}
	//昨日流量全天小时级访问统计
	yesterdayStat := []int64{}
	yesterdayTime := currentTime.Add(-1 * time.Duration(time.Hour*24))
	for i := 0; i <= 23; i++ {
		dateTime := time.Date(yesterdayTime.Year(), yesterdayTime.Month(), yesterdayTime.Day(), i, 0, 0, 0, lib.TimeLocation)
		hourData, _ := counter.GetHourData(dateTime)
		yesterdayStat = append(yesterdayStat, hourData)
	}
	stat := dto.StatisticsOutput{
		Today:     todayStat,
		Yesterday: yesterdayStat,
	}
	middleware.ResponseSuccess(c, stat)
	return
}

// AppAdd godoc
// @Summary 租户添加
// @Description 租户添加
// @Tags 租户管理
// @ID /app/app_add
// @Accept  json
// @Produce  json
// @Param body body dto.AppAddInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /app/app_add [post]
func (app *AppController) AppAdd(c *gin.Context) {
	// 获取添加租户输入结构体各项参数
	params := &dto.AppAddInput{}
	// 验证参数正确性
	if err := params.BindValidParam(c); err != nil {
		// 参数不正确返回错误信息并终止执行
		middleware.ResponseError(c, 2000, err)
		return
	}

	// 获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//验证app_id是否被占用
	search := &dao.App{
		AppID: params.AppID,
	}
	if _, err = search.FindFirst(c, tx, search); err == nil {
		middleware.ResponseError(c, 2002, errors.New("租户ID被占用，请重新输入"))
		return
	}
	// 如果没有主动设置密钥，则通过app_id用md5加密的方式生成密钥
	if params.Secret == "" {
		params.Secret = public.MD5(params.AppID)
	}
	// 组装appInfo
	// todo 结构体中需要添加时间属性
	appInfo := &dao.App{
		AppID:    params.AppID,
		Name:     params.Name,
		Secret:   params.Secret,
		WhiteIPS: params.WhiteIPS,
		Qps:      params.Qps,
		Qpd:      params.Qpd,
	}
	// 将添加租户数据表保存到数据库中
	if err = appInfo.Save(c, tx); err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, "")
	return
}

// AppUpdate godoc
// @Summary 租户更新
// @Description 租户更新
// @Tags 租户管理
// @ID /app/app_update
// @Accept  json
// @Produce  json
// @Param body body dto.AppUpdateInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /app/app_update [post]
func (app *AppController) AppUpdate(c *gin.Context) {
	// 获取修改租户输入结构体各项参数
	params := &dto.AppUpdateInput{}
	// 验证参数正确性
	if err := params.BindValidParam(c); err != nil {
		// 参数不正确返回错误信息并终止执行
		middleware.ResponseError(c, 2000, err)
		return
	}

	// 获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	// 通过租户ID来进行查询
	search := &dao.App{
		ID: params.ID,
	}

	// 获取租户详情信息
	appInfo, err := search.FindFirst(c, tx, search)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	// 如果没有主动设置密钥，则通过app_id用md5加密的方式生成密钥
	if params.Secret == "" {
		params.Secret = public.MD5(params.AppID)
	}
	// 根据传入的参数进行租户信息的修改
	appInfo.Name = params.Name
	appInfo.Secret = params.Secret
	appInfo.WhiteIPS = params.WhiteIPS
	appInfo.Qps = params.Qps
	appInfo.Qpd = params.Qpd

	// 将修改租户数据表保存到数据库中
	if err = appInfo.Save(c, tx); err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, "")
	return

}
