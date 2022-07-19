package dto

// PanelGroupDataOutput 面板分组数据结构体
type PanelGroupDataOutput struct {
	ServiceCount      int64 `json:"serviceCount"`
	AppCount          int64 `json:"appCount"`
	CurrentQPS        int64 `json:"currentQps"`
	TodayRequestCount int64 `json:"todayRequestCount"`
}

// FlowStatOutput 流量统计输出信息结构体
type FlowStatOutput struct {
	Today     []int64 `json:"today" form:"today" comment:"今日流量" example:"" validate:""`         //列表
	Yesterday []int64 `json:"yesterday" form:"yesterday" comment:"昨日流量" example:"" validate:""` //列表
}

// DashboardServiceStatItemOutput 首页大盘服务状态统计具体信息结构体
type DashboardServiceStatItemOutput struct {
	Name     string `json:"name"`
	LoadType int    `json:"load_type"`
	Value    int64  `json:"value"`
}

// DashboardServiceStatOutput 首页大盘服务状态统计信息结构体
type DashboardServiceStatOutput struct {
	Legend []string                         `json:"legend"`
	Data   []DashboardServiceStatItemOutput `json:"data"`
}
