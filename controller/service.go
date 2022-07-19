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
	"strings"
	"time"
)

type ServiceController struct{}

// ServiceRegister 注册服务列表信息相关路由
func ServiceRegister(group *gin.RouterGroup) {
	service := &ServiceController{}
	group.GET("/service_list", service.ServiceList)
	group.GET("/service_delete", service.ServiceDelete)
	group.GET("/service_detail", service.ServiceDetail)
	group.GET("/service_stat", service.ServiceStat)
	group.POST("/service_add_http", service.ServiceAddHTTP)
	group.POST("/service_update_http", service.ServiceUpdateHTTP)

	group.POST("/service_add_tcp", service.ServiceAddTcp)
	group.POST("/service_update_tcp", service.ServiceUpdateTcp)
	group.POST("/service_add_grpc", service.ServiceAddGrpc)
	group.POST("/service_update_grpc", service.ServiceUpdateGrpc)
}

// ServiceList godoc
// @Summary 服务列表
// @Description 服务列表
// @Tags 服务管理
// @ID /service/service_list
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param page_size query int true "每页个数"
// @Param page_no query int true "当前页数"
// @Success 200 {object} middleware.Response{data=dto.ServiceListOutput} "success"
// @Router /service/service_list [get]
func (service *ServiceController) ServiceList(c *gin.Context) {
	// 获取服务列表输入结构体各项参数
	params := &dto.ServiceListInput{}
	// 验证参数正确性
	if err := params.BindValidParam(c); err != nil {
		// 参数不正确返回错误信息并终止执行
		middleware.ResponseError(c, 2000, err)
		return
	}

	// 从数据库中读取ServiceInfo
	// 获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	// 从数据库中分页读取基本信息
	serviceInfo := &dao.ServiceInfo{}
	// 获取服务列表和服务总数
	list, total, err := serviceInfo.PageList(c, tx, params)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	// 根据返回的服务列表信息来组装返回值
	var outList []dto.ServiceListItemOutput
	for _, listItem := range list {
		// 获取服务详情信息
		serviceDetail, err := listItem.ServiceDetail(c, tx, &listItem)
		if err != nil {
			middleware.ResponseError(c, 2003, err)
			return
		}
		// 生成服务地址
		// 1.http类型地址生成的话：
		// 1.1. 前缀接入 clusterIP+clusterPort+path
		// 1.2. 域名接入 domain
		// 2.tcp、grpc接入 clusterIP+servicePort
		// cluster信息在conf.base.toml中定义
		serviceAddr := "unknow"
		clusterIP := lib.GetStringConf("base.cluster.cluster_ip")
		clusterPort := lib.GetStringConf("base.cluster.cluster_port")
		clusterSSLPort := lib.GetStringConf("base.cluster.cluster_ssl_port")

		// http类型且是后缀接入并且需要https则用clusterIP+clusterSSLPort+path生成服务地址
		if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL &&
			serviceDetail.HTTPRule.NeedHttps == 1 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterSSLPort, serviceDetail.HTTPRule.Rule)
		}
		// http类型且是前缀接入并且不需要https则用clusterIP+clusterPort+path生成服务地址
		if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL &&
			serviceDetail.HTTPRule.NeedHttps == 0 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterPort, serviceDetail.HTTPRule.Rule)
		}
		// http类型且是域名接入则用domain生成服务地址
		if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypeDomain {
			serviceAddr = serviceDetail.HTTPRule.Rule
		}

		// tcp类型则用clusterIP+servicePort生成服务地址
		if serviceDetail.Info.LoadType == public.LoadTypeTCP {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, serviceDetail.TCPRule.Port)
		}
		// grpc类型则用clusterIP+servicePort生成服务地址
		if serviceDetail.Info.LoadType == public.LoadTypeGRPC {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, serviceDetail.GRPCRule.Port)
		}
		// 获取所有的ip列表
		ipList := serviceDetail.LoadBalance.GetIPListByModel()
		// 获取服务计数器
		serviceCounter, err := public.FlowCounterHandler.GetCounter(public.FlowServicePrefix + listItem.ServiceName)
		if err != nil {
			middleware.ResponseError(c, 2004, err)
			return
		}
		// 组装服务的详情信息
		outItem := dto.ServiceListItemOutput{
			ID:          listItem.ID,
			LoadType:    listItem.LoadType,
			ServiceName: listItem.ServiceName,
			ServiceDesc: listItem.ServiceDesc,
			ServiceAddr: serviceAddr,
			Qps:         serviceCounter.QPS,
			Qpd:         serviceCounter.TotalCount,
			TotalNode:   len(ipList),
		}
		// 组装返回信息
		outList = append(outList, outItem)
	}
	// 组装返回信息
	out := &dto.ServiceListOutput{
		Total: total,
		List:  outList,
	}
	middleware.ResponseSuccess(c, out)
}

// ServiceDelete godoc
// @Summary 服务删除
// @Description 服务删除
// @Tags 服务管理
// @ID /service/service_delete
// @Accept  json
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_delete [get]
func (service *ServiceController) ServiceDelete(c *gin.Context) {
	// 获取服务删除输入结构体各项参数
	params := &dto.ServiceDeleteInput{}
	// 验证参数正确性
	if err := params.BindValidParam(c); err != nil {
		// 参数不正确返回错误信息并终止执行
		middleware.ResponseError(c, 2000, err)
		return
	}

	// 从数据库中读取ServiceInfo
	// 获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	// 从数据库中分页读取基本信息
	serviceInfo := &dao.ServiceInfo{ID: params.ID}
	// 获取服务信息
	serviceInfo, err = serviceInfo.Find(c, tx, serviceInfo)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	// 将服务信息中的IsDelete置为1，进行软删除
	serviceInfo.IsDelete = 1
	// 保存到数据库
	if err = serviceInfo.Save(c, tx); err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, "")
}

// ServiceAddHTTP godoc
// @Summary 添加HTTP服务
// @Description 添加HTTP服务
// @Tags 服务管理
// @ID /service/service_add_http
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddHTTPInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_add_http [post]
func (service *ServiceController) ServiceAddHTTP(c *gin.Context) {
	// 获取添加HTTP服务时输入的各项参数
	params := &dto.ServiceAddHTTPInput{}
	// 验证参数正确性
	if err := params.BindValidParam(c); err != nil {
		// 参数不正确返回错误信息并终止执行
		middleware.ResponseError(c, 2000, err)
		return
	}

	// ip列表数量与权重列表数量要相同
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2004, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	//从数据库中读取ServiceInfo
	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	// 开启数据库事务
	tx = tx.Begin()

	// 通过服务名查询数据库
	serviceInfo := &dao.ServiceInfo{ServiceName: params.ServiceName}
	// 如果err==nil，说明数据库中有同名服务
	if _, err = serviceInfo.FindFirst(c, tx, serviceInfo); err == nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2002, errors.New("服务已存在"))
		return
	}
	// 按照传入的 params.RuleType 来选择写入到数据库中是前缀接入还是域名接入
	httpUrl := &dao.HttpRule{RuleType: params.RuleType, Rule: params.Rule}
	// 如果err==nil，说明数据库中有相同的前缀或者域名
	if _, err = httpUrl.FindFirst(c, tx, httpUrl); err == nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2003, errors.New("服务接入前缀或域名已存在"))
		return
	}

	// 创建服务信息数据表
	serviceModel := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	// 将服务信息数据表保存到数据库中
	if err = serviceModel.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2005, err)
		return
	}

	// 创建httpRule信息数据表
	httpRule := &dao.HttpRule{
		ServiceID:      serviceModel.ID,
		RuleType:       params.RuleType,
		Rule:           params.Rule,
		NeedHttps:      params.NeedHttps,
		NeedStripUri:   params.NeedStripUri,
		NeedWebsocket:  params.NeedWebsocket,
		UrlRewrite:     params.UrlRewrite,
		HeaderTransfor: params.HeaderTransfor,
	}
	// 将httpRule信息数据表保存到数据库中
	if err = httpRule.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}

	// 创建权限控制信息数据表
	accessControl := &dao.AccessControl{
		ServiceID:         serviceModel.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		ClientIPFlowLimit: params.ClientipFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}

	// 将权限控制信息数据表保存到数据库中
	if err = accessControl.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}

	// 创建负载均衡信息数据表
	loadBalance := &dao.LoadBalance{
		ServiceID:              serviceModel.ID,
		RoundType:              params.RoundType,
		IpList:                 params.IpList,
		WeightList:             params.WeightList,
		UpstreamConnectTimeout: params.UpstreamConnectTimeout,
		UpstreamIdleTimeout:    params.UpstreamIdleTimeout,
		UpstreamMaxIdle:        params.UpstreamMaxIdle,
	}
	// 将负载均衡信息数据表保存到数据库中
	if err = loadBalance.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}
	// 提交事务，数据入库
	tx.Commit()
	middleware.ResponseSuccess(c, "")
}

// ServiceUpdateHTTP godoc
// @Summary 修改HTTP服务
// @Description 修改HTTP服务
// @Tags 服务管理
// @ID /service/service_update_http
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateHTTPInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_update_http [post]
func (service *ServiceController) ServiceUpdateHTTP(c *gin.Context) {
	// 获取修改HTTP服务时输入的各项参数
	params := &dto.ServiceUpdateHTTPInput{}
	// 验证参数正确性
	if err := params.BindValidParam(c); err != nil {
		// 参数不正确返回错误信息并终止执行
		middleware.ResponseError(c, 2000, err)
		return
	}

	// ip列表数量与权重列表数量要相同
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2001, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	//从数据库中读取ServiceInfo
	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	// 开启数据库事务
	tx = tx.Begin()
	// 通过服务名称查询数据库
	serviceInfo := &dao.ServiceInfo{ServiceName: params.ServiceName}
	serviceInfo, err = serviceInfo.FindFirst(c, tx, serviceInfo)
	if err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2003, errors.New("服务不存在"))
		return
	}

	//获取服务详细信息
	serviceDetail, err := serviceInfo.ServiceDetail(c, tx, serviceInfo)
	if err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2004, errors.New("服务不存在"))
		return
	}

	// 将更新后的服务信息保存到数据库中
	info := serviceDetail.Info
	info.ServiceDesc = params.ServiceDesc
	if err = info.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2005, err)
		return
	}

	// 更新httpRule信息并保存到数据库中
	httpRule := serviceDetail.HTTPRule
	httpRule.NeedHttps = params.NeedHttps
	httpRule.NeedStripUri = params.NeedStripUri
	httpRule.NeedWebsocket = params.NeedWebsocket
	httpRule.UrlRewrite = params.UrlRewrite
	httpRule.HeaderTransfor = params.HeaderTransfor
	if err = httpRule.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return

	}

	// 更新权限控制信息并保存到数据库中
	accessControl := serviceDetail.AccessControl
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.ClientIPFlowLimit = params.ClientipFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err = accessControl.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2007, err)
		return

	}

	// 更新负载均衡信息并保存到数据库中
	loadBalance := serviceDetail.LoadBalance
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.UpstreamConnectTimeout = params.UpstreamConnectTimeout
	loadBalance.UpstreamHeaderTimeout = params.UpstreamHeaderTimeout
	loadBalance.UpstreamIdleTimeout = params.UpstreamIdleTimeout
	loadBalance.UpstreamMaxIdle = params.UpstreamMaxIdle
	if err = loadBalance.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}
	// 提交事务，数据入库
	tx.Commit()
	middleware.ResponseSuccess(c, "")
}

// ServiceDetail godoc
// @Summary 服务详情
// @Description 服务详情
// @Tags 服务管理
// @ID /service/service_detail
// @Accept  json
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} middleware.Response{data=dao.ServiceDetail} "success"
// @Router /service/service_detail [get]
func (service *ServiceController) ServiceDetail(c *gin.Context) {
	// 获取服务具体输入结构体各项参数
	params := &dto.ServiceDetailInput{}
	// 验证参数正确性
	if err := params.BindValidParam(c); err != nil {
		// 参数不正确返回错误信息并终止执行
		middleware.ResponseError(c, 2000, err)
		return
	}

	// 从数据库中读取ServiceInfo
	// 获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	// 从数据库中分页读取基本信息
	serviceInfo := &dao.ServiceInfo{ID: params.ID}
	// 获取服务信息
	serviceInfo, err = serviceInfo.Find(c, tx, serviceInfo)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	// 获取服务详情信息
	serviceDetail, err := serviceInfo.ServiceDetail(c, tx, serviceInfo)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, serviceDetail)
}

// ServiceStat godoc
// @Summary 服务统计
// @Description 服务统计
// @Tags 服务管理
// @ID /service/service_stat
// @Accept  json
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} middleware.Response{data=dto.ServiceStatOutput} "success"
// @Router /service/service_stat [get]
func (service *ServiceController) ServiceStat(c *gin.Context) {
	// 获取服务统计输入结构体各项参数
	params := &dto.ServiceStatInput{}
	// 验证参数正确性
	if err := params.BindValidParam(c); err != nil {
		// 参数不正确返回错误信息并终止执行
		middleware.ResponseError(c, 2000, err)
		return
	}

	// 从数据库中读取ServiceInfo
	// 获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	// 从数据库中分页读取基本信息
	serviceInfo := &dao.ServiceInfo{ID: params.ID}
	// 获取服务信息
	serviceInfo, err = serviceInfo.Find(c, tx, serviceInfo)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	// 获取服务详情信息
	serviceDetail, err := serviceInfo.ServiceDetail(c, tx, serviceInfo)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}

	counter, err := public.FlowCounterHandler.GetCounter(public.FlowServicePrefix + serviceDetail.Info.ServiceName)
	if err != nil {
		middleware.ResponseError(c, 2004, err)
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

// ServiceAddTcp godoc
// @Summary TCP服务添加
// @Description TCP服务添加
// @Tags 服务管理
// @ID /service/service_add_tcp
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddTcpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_add_tcp [post]
func (service *ServiceController) ServiceAddTcp(c *gin.Context) {
	// 获取添加TCP服务时输入的各项参数
	params := &dto.ServiceAddTcpInput{}
	// 验证参数正确性
	if err := params.BindValidParam(c); err != nil {
		// 参数不正确返回错误信息并终止执行
		middleware.ResponseError(c, 2000, err)
		return
	}

	// 从数据库中读取ServiceInfo
	// 获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	infoSearch := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		IsDelete:    0,
	}
	// 验证serviceName是否被占用过
	if _, err = infoSearch.FindFirst(c, tx, infoSearch); err == nil {
		middleware.ResponseError(c, 2002, errors.New("服务名被占用，请重新输入"))
		return
	}

	// 验证端口是否被占用
	// 首先验证tcp中是否有该端口
	tcpRuleSearch := &dao.TcpRule{
		Port: params.Port,
	}
	if _, err = tcpRuleSearch.FindFirst(c, tx, tcpRuleSearch); err == nil {
		middleware.ResponseError(c, 2003, errors.New("服务端口被占用，请重新输入"))
		return
	}
	// 再验证grpc中是否有该端口
	grpcRuleSearch := &dao.GrpcRule{
		Port: params.Port,
	}
	if _, err = grpcRuleSearch.FindFirst(c, tx, grpcRuleSearch); err == nil {
		middleware.ResponseError(c, 2004, errors.New("服务端口被占用，请重新输入"))
		return
	}

	// ip列表数量与权重列表数量要相同
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2005, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	// 开启数据库事务
	tx = tx.Begin()

	// 创建服务信息数据表
	info := &dao.ServiceInfo{
		LoadType:    public.LoadTypeTCP,
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	// 将服务信息数据表保存到数据库中
	if err = info.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}

	// 创建负载均衡信息数据表
	loadBalance := &dao.LoadBalance{
		ServiceID:  info.ID,
		RoundType:  params.RoundType,
		IpList:     params.IpList,
		WeightList: params.WeightList,
		ForbidList: params.ForbidList,
	}
	// 将负载均衡信息数据表保存到数据库中
	if err = loadBalance.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}

	// 创建tcpRule信息数据表
	tcpRule := &dao.TcpRule{
		ServiceID: info.ID,
		Port:      params.Port,
	}
	// 将tcpRule信息数据表保存到数据库中
	if err = tcpRule.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}
	// 创建权限控制信息数据表
	accessControl := &dao.AccessControl{
		ServiceID:         info.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		WhiteHostName:     params.WhiteHostName,
		ClientIPFlowLimit: params.ClientIPFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	// 将权限控制信息数据表保存到数据库中
	if err = accessControl.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2009, err)
		return
	}
	// 提交事务，数据入库
	tx.Commit()
	middleware.ResponseSuccess(c, "")
}

// ServiceUpdateTcp godoc
// @Summary 修改Tcp服务
// @Description 修改Tcp服务
// @Tags 服务管理
// @ID /service/service_update_tcp
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateTcpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_update_tcp [post]
func (service *ServiceController) ServiceUpdateTcp(c *gin.Context) {
	// 获取修改Tcp服务时输入的各项参数
	params := &dto.ServiceUpdateTcpInput{}
	// 验证参数正确性
	if err := params.BindValidParam(c); err != nil {
		// 参数不正确返回错误信息并终止执行
		middleware.ResponseError(c, 2000, err)
		return
	}

	// ip列表数量与权重列表数量要相同
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2001, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	//从数据库中读取ServiceInfo
	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	// 开启数据库事务
	tx = tx.Begin()
	// 通过服务ID查询数据库
	serviceInfo := &dao.ServiceInfo{ID: params.ID}

	//获取服务详细信息
	serviceDetail, err := serviceInfo.ServiceDetail(c, tx, serviceInfo)
	if err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2003, err)
		return
	}

	// 将更新后的服务信息保存到数据库中
	info := serviceDetail.Info
	info.ServiceDesc = params.ServiceDesc
	if err = info.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2004, err)
		return
	}

	// 更新tcpRule信息并保存到数据库中
	tcpRule := &dao.TcpRule{}
	if serviceDetail.TCPRule != nil {
		tcpRule = serviceDetail.TCPRule
	}
	tcpRule.ServiceID = info.ID
	tcpRule.Port = params.Port
	if err = tcpRule.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2005, err)
		return
	}
	// 更新权限控制信息并保存到数据库中
	accessControl := &dao.AccessControl{}
	if serviceDetail.AccessControl != nil {
		accessControl = serviceDetail.AccessControl
	}
	accessControl.ServiceID = info.ID
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.WhiteHostName = params.WhiteHostName
	accessControl.ClientIPFlowLimit = params.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err = accessControl.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}

	// 更新负载均衡信息并保存到数据库中
	loadBalance := &dao.LoadBalance{}
	if serviceDetail.LoadBalance != nil {
		loadBalance = serviceDetail.LoadBalance
	}
	loadBalance.ServiceID = info.ID
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.ForbidList = params.ForbidList
	if err = loadBalance.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}

	// 提交事务，数据入库
	tx.Commit()
	middleware.ResponseSuccess(c, "")
}

// ServiceAddGrpc godoc
// @Summary grpc服务添加
// @Description grpc服务添加
// @Tags 服务管理
// @ID /service/service_add_grpc
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddGrpcInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_add_grpc [post]
func (service *ServiceController) ServiceAddGrpc(c *gin.Context) {
	// 获取添加Grpc服务时输入的各项参数
	params := &dto.ServiceAddGrpcInput{}
	// 验证参数正确性
	if err := params.BindValidParam(c); err != nil {
		// 参数不正确返回错误信息并终止执行
		middleware.ResponseError(c, 2000, err)
		return
	}

	// 从数据库中读取ServiceInfo
	// 获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	infoSearch := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		IsDelete:    0,
	}
	// 验证serviceName是否被占用过
	if _, err = infoSearch.FindFirst(c, tx, infoSearch); err == nil {
		middleware.ResponseError(c, 2002, errors.New("服务名被占用，请重新输入"))
		return
	}

	// 验证端口是否被占用
	// 首先验证tcp中是否有该端口
	tcpRuleSearch := &dao.TcpRule{
		Port: params.Port,
	}
	if _, err = tcpRuleSearch.FindFirst(c, tx, tcpRuleSearch); err == nil {
		middleware.ResponseError(c, 2003, errors.New("服务端口被占用，请重新输入"))
		return
	}
	// 再验证grpc中是否有该端口
	grpcRuleSearch := &dao.GrpcRule{
		Port: params.Port,
	}
	if _, err = grpcRuleSearch.FindFirst(c, tx, grpcRuleSearch); err == nil {
		middleware.ResponseError(c, 2004, errors.New("服务端口被占用，请重新输入"))
		return
	}

	// ip列表数量与权重列表数量要相同
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2005, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	// 开启数据库事务
	tx = tx.Begin()

	// 创建服务信息数据表
	info := &dao.ServiceInfo{
		LoadType:    public.LoadTypeGRPC,
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	// 将服务信息数据表保存到数据库中
	if err = info.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}

	// 创建负载均衡信息数据表
	loadBalance := &dao.LoadBalance{
		ServiceID:  info.ID,
		RoundType:  params.RoundType,
		IpList:     params.IpList,
		WeightList: params.WeightList,
		ForbidList: params.ForbidList,
	}
	// 将负载均衡信息数据表保存到数据库中
	if err = loadBalance.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}

	// 创建grpcRule信息数据表
	grpcRule := &dao.GrpcRule{
		ServiceID:      info.ID,
		Port:           params.Port,
		HeaderTransfor: params.HeaderTransfor,
	}
	// 将grpcRule信息数据表保存到数据库中
	if err = grpcRule.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}
	// 创建权限控制信息数据表
	accessControl := &dao.AccessControl{
		ServiceID:         info.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		WhiteHostName:     params.WhiteHostName,
		ClientIPFlowLimit: params.ClientIPFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	// 将权限控制信息数据表保存到数据库中
	if err = accessControl.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2009, err)
		return
	}
	// 提交事务，数据入库
	tx.Commit()
	middleware.ResponseSuccess(c, "")
}

// ServiceUpdateGrpc godoc
// @Summary 修改Grpc服务
// @Description 修改Grpc服务
// @Tags 服务管理
// @ID /service/service_update_grpc
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateGrpcInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_update_grpc [post]
func (service *ServiceController) ServiceUpdateGrpc(c *gin.Context) {
	// 获取修改Grpc服务时输入的各项参数
	params := &dto.ServiceUpdateGrpcInput{}
	// 验证参数正确性
	if err := params.BindValidParam(c); err != nil {
		// 参数不正确返回错误信息并终止执行
		middleware.ResponseError(c, 2000, err)
		return
	}

	// ip列表数量与权重列表数量要相同
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2001, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	//从数据库中读取ServiceInfo
	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	// 开启数据库事务
	tx = tx.Begin()
	// 通过服务ID查询数据库
	serviceInfo := &dao.ServiceInfo{ID: params.ID}

	//获取服务详细信息
	serviceDetail, err := serviceInfo.ServiceDetail(c, tx, serviceInfo)
	if err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2003, err)
		return
	}

	// 将更新后的服务信息保存到数据库中
	info := serviceDetail.Info
	info.ServiceDesc = params.ServiceDesc
	if err = info.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2004, err)
		return
	}

	// 更新tcpRule信息并保存到数据库中
	grpcRule := &dao.GrpcRule{}
	if serviceDetail.GRPCRule != nil {
		grpcRule = serviceDetail.GRPCRule
	}
	grpcRule.ServiceID = info.ID
	grpcRule.Port = params.Port
	grpcRule.HeaderTransfor = params.HeaderTransfor
	if err = grpcRule.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2005, err)
		return
	}
	// 更新权限控制信息并保存到数据库中
	accessControl := &dao.AccessControl{}
	if serviceDetail.AccessControl != nil {
		accessControl = serviceDetail.AccessControl
	}
	accessControl.ServiceID = info.ID
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.WhiteHostName = params.WhiteHostName
	accessControl.ClientIPFlowLimit = params.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err = accessControl.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}

	// 更新负载均衡信息并保存到数据库中
	loadBalance := &dao.LoadBalance{}
	if serviceDetail.LoadBalance != nil {
		loadBalance = serviceDetail.LoadBalance
	}
	loadBalance.ServiceID = info.ID
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.ForbidList = params.ForbidList
	if err = loadBalance.Save(c, tx); err != nil {
		// 出现err就回滚事务
		tx.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}

	// 提交事务，数据入库
	tx.Commit()
	middleware.ResponseSuccess(c, "")
}
