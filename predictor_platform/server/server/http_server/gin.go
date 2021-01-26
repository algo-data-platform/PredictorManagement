package http_server

import (
	"github.com/DeanThompson/ginpprof"
	"github.com/gin-gonic/gin"
	"server/env"
	"server/libs/logger"
	"server/metrics"
	"server/server/http_server/api"
	"strconv"
)

// 配置gin路由
func setRouter(router *gin.Engine) {
	// static router
	if env.Env.Conf.RunEnv == "prod" {
		router.LoadHTMLGlob("frontend/dist/*.html") // add entrance of index.html
		router.Static("/static", "frontend/dist/static")
		router.StaticFile("/", "frontend/dist/index.html")
	}

	// mysql 数据表操作
	router.GET("/mysql/tables", api.MysqlTables)
	router.GET("/mysql/show", api.MysqlShow)
	router.GET("/mysql/insert", api.MysqlInsert)
	router.GET("/mysql/delete", api.MysqlDelete)
	router.GET("/mysql/update", api.MysqlUpdate)
	router.POST("/mysql/batch_insert_hosts", api.MysqlBatchInsertHosts)

	// node_info 机器资源监控信息
	router.GET("/node_infos", api.NodeInfos)

	// will refactor this interface later
	router.GET("/user/login", api.UserLogin)
	router.GET("/user/logout", api.UserLogout)

	// 模型信息相关操作
	router.GET("/model_info/model_history", api.ModelInfoModelHistory)
	router.GET("/model_info/update_interval_week", api.ModelInfoUpdateIntervalWeek)
	router.GET("/model_info/models_mail_recipients", api.ModelInfoModelsMailRecipients)
	router.GET("/model_info/set_model_mail_recipients", api.ModelInfoSetModelMailRecipients)

	// 动态扩容相关操作
	router.GET("/elastic_expansion/insert", api.InsertElasticExpansion)
	router.GET("/elastic_expansion/delete", api.DeleteElasticExpansion)

	// load balance group
	router.GET("/load_balance/insert", api.InsertLoadBalance)
	router.GET("/load_balance/get", api.GetLoadBalance)
	router.GET("/load_balance/delete", api.DeleteLoadBalance)
	router.GET("/load_balance/reset", api.ResetLoadBalance)
	router.GET("/load_balance/reset_weight", api.ResetLoadWeight)

	// grafana webhook group
	router.POST("/webhook/alert", api.AlertWebhook)

	// add prometheus router
	reporter := metrics.GetReporter()
	router.GET("/server/prometheus", gin.WrapH(reporter.HTTPHandler()))
	router.GET("/server/json", gin.WrapH(reporter.JsonHTTPHandler()))

	// 预测服务召回http接口
	router.POST("/predictor/calculate_vector", api.PredictorCalculateVector)

	// 自动压测相关
	router.GET("/stress/insert", api.StressInsert)
	router.GET("/stress/disable", api.StressDisable)
	router.GET("/stress/list", api.StressList)
	router.GET("/stress/enable", api.StressEnable)
	router.GET("/stress/save_qps", api.StressSaveQps)

	// 降级
	router.GET("/downgrade/get_prometheus_downgrade_percent", api.GetPromDowngradePercent)
	router.GET("/downgrade/set_by_service", api.SetDowngradeByService)
	router.GET("/downgrade/reset_by_service", api.ResetDowngradeByService)

	// 半自动迁移机器
	migrate := &api.Migrate{}
	router.GET("/migrate/service_stats", migrate.GetServiceStats)
	router.GET("/migrate/get_from_services", migrate.GetFromServices)
	router.GET("/migrate/get_to_services", migrate.GetToServices)
	router.GET("/migrate/preview", migrate.Preview)
	router.GET("/migrate/do_migrate", migrate.DoMigrate)
}

// 增加跨域header
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		// in case of cors
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers",
			"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	}
}

// gin http server start
func Start() {
	gin.SetMode(gin.ReleaseMode)
	// redirect gin log
	gin.DefaultWriter = env.Env.LogWriter
	gin.DefaultErrorWriter = env.Env.LogWriter
	router := gin.Default()
	router.Use(Cors())
	setRouter(router)
	// 添加服务心跳metirc
	metrics.GetGauges()[metrics.GAUGE_SERVING].Update(1)
	host_port := (env.Env.Conf.HttpHost) + ":" + strconv.Itoa(env.Env.Conf.HttpPort)
	logger.Infof("server address: %s", host_port)
	ginpprof.Wrapper(router)
	err := router.Run(host_port)
	if err != nil {
		logger.Fatalf("router.Run error: %v", err)
	}
}
