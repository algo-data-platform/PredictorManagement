package server

import (
	"server/common"
	"server/env"
	"server/metrics"
	"server/server/dao"
	"server/server/http_server"
	"server/server/service"
)

func Init() {
	common.GLoadThresholdServices = []string{}
	dao.SetMysqlDB(env.Env.MysqlDB)
	dao.TableCheck()
	metrics.InitMetrics()
}

// 启动服务
func startService() {
	// 第一次启动初始化nodeInfos
	service.UpdateNodeInfos(env.Env.Conf)
	go service.TimeUpdateNodeInfos(env.Env.Conf)
	// 模型过旧监控，模型加载状态监控
	go service.StartMonitor(env.Env.Conf)
	// 判断是否只在cron机器运行
	if env.Env.LocalIp == env.Env.Conf.CronTabIp {
		// 模型时效性邮件定时发送
		go service.ModelTimeMail(env.Env.Conf)
		// 自动权重
		go service.UpdateCpuLoad(env.Env.Conf)
		// 开启动态扩容机器分配脚本
		go service.CheckElasticExpansion(env.Env.Conf)
		// consul fallback，定时同步静态ip列表
		go service.GeneratePredictorStaticIpList(env.Env.Conf)
	}
}

func Start() {
	// 初始化
	Init()
	// 启动服务
	startService()
	// 启动gin http服务
	http_server.Start()
}
