package util

import (
	"github.com/algo-data-platform/predictor/golibs/adgo/predictor_client/if/predictor"
)

type NodeOn int

const (
	Node_Off NodeOn = iota
	Node_On
)

//easyjson:json
type LoadAverageAll struct {
	Average1  float32
	Average5  float32
	Average15 float32
}

//easyjson:json
type NodeResourceInfo struct {
	CoreNum                 int
	TotalMem                int64
	AvailMem                int64
	TotalDisk               int64
	AvailDisk               int64
	LoadAverage             LoadAverageAll
	NodeAvail               NodeOn //根据心跳判断，放在resource struct
	LastCpuIdleSecondsTotal int64
	LastUptime              int64
	Cpu                     float64
}

//easyjson:json
type NodeInfo struct {
	Host         string           // 从mysql的hosts列表中读取
	DataCenter   string           //host所在的data center
	ResourceInfo NodeResourceInfo // 对应9100:metrics
	StatusInfo   []ServiceInfo    // 对应url如http://10.85.101.119:9538/get_service_model_info
}

//mysql table: services
type Services struct {
	ServiceId   int
	ServiceName string
	Description string
}

type ModelInfo struct {
	SuccessTime string `json:"success_time"`
	FullName    string `json:"fullName"`
	ConfigName  string `json:"configName"`
	Name        string `json:"name"`
	Timestamp   string `json:"timestamp"`
	Md5         string `json:"md5"`
	State       string `json:"state"`
}

//mysql show all tablenams
type AllTable struct {
	Database   string
	TableNames []string
}

type ServiceInfo struct {
	ModelRecords  []ModelInfo `json:"model_records"`
	ServiceName   string      `json:"service_name"`
	ServiceWeight float64     `json:"service_weight"`
}

//add mock data struct of data_source
type NodeMetaInfo struct {
	Msg struct {
		Services []ServiceInfo `json:"services"`
	} `json:"msg"`
	Code int `json:"code"`
}

type ModelHistoryInfo struct {
	ModelName string
	Timestamp string
	Status    string
	Percent   int
	Md5       string
	IsLocked  uint
	Desc      string
	CreatedAt string
	UpdatedAt string
}

// add 模型时效性
type ModelUpdateTimingInfo struct {
	ModelName             string
	ModelUpdateTimeWeekly int64 // 模型一周内更新平均耗时
	LastestTimestampArray []string
	ModelChannel          string
}

// 模型版本信息
type ModelVersionInfo struct {
	ModelName      string
	ModelTimestamp string
}

type IP_Service struct {
	IP             string
	Service        string
	Service_weight uint
}

type ServiceWeight struct {
	Service_weight uint
	Cpu_use        float64
}

type UpdateInfo struct {
	Hid  uint
	Sid  uint
	Load uint
}

// 针对http请求的json结构返回
type JsonRespData struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

// grafana webhook 请求结构
type WebHookRequest struct {
	Title       string      `json:"title"`
	Tags        interface{} `json:"tags"`
	State       string      `json:"state"`
	RuleUrl     string      `json:"ruleUrl"`
	RuleName    string      `json:"ruleName"`
	RuleId      int         `json:"ruleId"`
	ImageUrl    string      `json:"imageUrl"`
	EvalMatches []EvalMatch `json:"evalMatches"`
}

type EvalMatch struct {
	Value  interface{} `json:"value"`
	Metric string      `json:"metric"`
	Tags   interface{} `json:"tags"`
}

type ModelExtension struct {
	Name      string
	Extension string
}

type ModelExtensionInfo struct {
	MailRecipients []string `json:MailRecipients`
}

type SidHostNum struct {
	Sid     uint
	HostNum int
}

type IPWeight struct {
	HostIp     string
	LoadWeight int
}

type CalculateRequest struct {
	Reqs        []*predictor.CalculateVectorRequest `json:"reqs"`
	ServiceName string                              `json:"service_name"`
	TimeoutMS   int                                 `json:"timeout_ms"`
}

type StressInfo struct {
	ID         uint     `json:"id"`
	Hid        uint     `json:"hid"`
	IP         string   `json:"ip"`
	Mids       []uint   `json:"mids"`
	QPS        []uint   `json:"qps"`
	ModelNames []string `json:"model_names"`
	CreateTime string   `json:"create_time"`
	UpdateTime string   `json:"update_time"`
	IsEnable   uint     `json:"is_enable"`
}

type VersionStatus struct {
	Timestamp string
	State     string
}

type DowngradeResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type IpSid struct {
	Ip  string
	Sid uint
}

type SidName struct {
	Sids    []uint   `json:"sids"`
	Names   []string `json:"names"`
	HostNum int      `json:"host_num"`
}

type PreviewHost struct {
	Hsid         uint             `json:"hsid"`
	Hid          uint             `json:"hid"`
	Ip           string           `json:"ip"`
	Sids         []uint           `json:"sids"`
	ServiceNames []string         `json:"service_names"`
	ResourceInfo NodeResourceInfo `json:"resource_info"`
	IDC          string           `json:"idc"`
}

type HostServiceInfo struct {
	Hsid uint   `json:"hsid"`
	Hid  uint   `json:"hid"`
	Ip   string `json:"ip"`
	Sid  uint   `json:"sid"`
}

type ServiceStats struct {
	Sid         uint         `json:"sid"`
	ServiceName string       `json:"service_name"`
	HostNum     int          `json:"host_num"`
	IDCHostNums []IDCHostNum `json:"idc_host_nums"`
	CpuHostNums []CpuHostNum `json:"cpu_host_nums"`
	MemHostNums []MemHostNum `json:"mem_host_nums"`
}

type IDCHostNum struct {
	IDC     string `json:"idc"`
	HostNum int    `json:"host_num"`
}

type CpuHostNum struct {
	CoreNum int `json:"core_num"`
	HostNum int `json:"host_num"`
}

type MemHostNum struct {
	TotalMem int `json:"total_mem"`
	HostNum  int `json:"host_num"`
}

type GroupSid []uint
