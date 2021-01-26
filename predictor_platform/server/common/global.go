package common

import (
	"server/util"
)

// 全局变量定义，前缀G
// load_balance 开启自动权重的service配置
var GLoadThresholdServices []string

// 全局资源信息变量
var GNodeInfos []util.NodeInfo

// ip段对应机房
var GIpToIDCMap = map[string]string{
	"127.2": "dbl",
	"127.1": "huawei",
	"127.4": "aliyun",
	"127.3": "bx",
}
