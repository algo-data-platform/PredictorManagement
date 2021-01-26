package conf

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"path"
	"path/filepath"
	"server/libs/logger"
	"strconv"
	"strings"
)

type Log struct {
	FileName       string
	Level          string
	IsRotate       bool
	RotateCycle    string
	RotateMaxHours int
}

type Mysql struct {
	Driver     string
	Host       string
	Port       string
	User       string
	Passwd     string
	Database   string
	TableNames []string
	AlarmList  []string
}

type Monitor struct {
	CheckInterval    int
	ExcludedServices []string
}

type ModelTransmit struct {
	SrcHost string
	SrcPath string
}
type ElasticExpansion struct {
	CheckInterval int
	ServiceGroups []ServiceGroup
}

type ServiceGroup struct {
	Services []string
}

type LoadThreshold struct {
	CpuLimit      float64
	CheckInterval int
	Method        string
	Up_Gap        int
	Down_Gap      int
}

type ServiceStaticList struct {
	UpdateInterval     int
	AdServerRelayHost  string
	PredictorRelayHost string
	SubDataDir         string
}

type PredictorClient struct {
	ConsulAddress string
}

type Prometheus struct {
	Address string
}

type MigrateHosts struct {
	ExcludeHosts []string
}

type Conf struct {
	Log                Log
	HttpHost           string
	HttpPort           int
	HttpTimeout        int
	HttpRetryConn      int
	PredictorHttpPort  int
	ModelTimingRange   int
	UsernameEmail      string
	PasswdEmail        string
	HostEmail          string
	PortEmail          int
	HtmlTemplatePath   string
	FromEmail          string
	EmailCronTime      string
	CronTabIp          string
	Recipients         []string
	MysqlDb            Mysql
	Monitor            Monitor
	ModelChannels      []string
	ModelTransmit      ModelTransmit
	RunEnv             string
	LoadThreshold      LoadThreshold
	ElasticExpansion   ElasticExpansion
	DataDir            string
	ServiceStaticList  ServiceStaticList
	PredictorClient    PredictorClient
	StressTestService  string
	ConsistenceService string
	Prometheus         Prometheus
	MigrateHosts       MigrateHosts
}

var conf *Conf
var DestPath string

func New() {
	conf = &Conf{}
	var configFile string
	flag.StringVar(&configFile, "conf", "./conf/config.json", "config file prefix")
	flag.Parse()
	configPath, fileName := filepath.Split(configFile)
	fileExt := path.Ext(fileName)
	configName := strings.TrimSuffix(fileName, fileExt)
	viper.SetConfigName(configName)
	viper.AddConfigPath(configPath)
	viper.SetConfigType("json")
	err := viper.ReadInConfig()
	if err != nil {
		logger.Fatalf("ReadInConfig err: %v", err)
		return
	}

	http_host := fmt.Sprintf("%v", viper.Get("http_host"))
	http_port_str := fmt.Sprintf("%v", viper.Get("http_port"))
	http_timeout_str := fmt.Sprintf("%v", viper.Get("http_timeout"))
	http_retry_conn_str := fmt.Sprintf("%v", viper.Get("http_retry_conn"))
	predictor_http_port_str := fmt.Sprintf("%v", viper.Get("predictor_http_port"))
	model_timeing_range_str := fmt.Sprintf("%v", viper.Get("model_timing_range"))
	recipients_list_str := fmt.Sprintf("%v", viper.Get("recipients"))
	recipients_list := strings.Split(strings.Trim(recipients_list_str, "[]"), " ")

	// log
	conf.Log.FileName = viper.GetString("log.fileName")
	conf.Log.Level = viper.GetString("log.level")
	conf.Log.IsRotate = viper.GetBool("log.isRotate")
	conf.Log.RotateCycle = viper.GetString("log.rotateCycle")
	conf.Log.RotateMaxHours = viper.GetInt("log.rotateMaxHours")
	conf.HttpHost = http_host
	port, err := strconv.Atoi(http_port_str)
	if err != nil {
		logger.Fatalf("strconv error: %v", err)
		return
	}
	conf.HttpPort = port
	timeout, err := strconv.Atoi(http_timeout_str)
	if err != nil {
		logger.Fatalf("strconv error: %v", err)
		return
	}
	conf.HttpTimeout = timeout
	http_retry_conn, err := strconv.Atoi(http_retry_conn_str)
	if err != nil {
		logger.Fatalf("strconv error: %v", err)
		return
	}
	conf.HttpRetryConn = http_retry_conn
	predictor_http_port, err := strconv.Atoi(predictor_http_port_str)
	if err != nil {
		logger.Fatalf("strconv error: %v", err)
		return
	}
	model_timing_range, err := strconv.Atoi(model_timeing_range_str)
	if err != nil {
		logger.Fatalf("strconv error: %v", err)
		return
	}
	conf.PredictorHttpPort = predictor_http_port
	conf.ModelTimingRange = model_timing_range
	conf.UsernameEmail = viper.GetString("username_email")
	conf.PasswdEmail = viper.GetString("passwd_email")
	conf.HostEmail = viper.GetString("host_email")
	conf.FromEmail = viper.GetString("from_email")
	conf.HtmlTemplatePath = viper.GetString("html_template_path")
	conf.EmailCronTime = viper.GetString("email_cron_time")
	conf.CronTabIp = viper.GetString("cron_tab_ip")
	conf.Recipients = recipients_list
	// mysql conf init
	mysql_driver_str := viper.Get("mysql.driver")
	mysql_driver := fmt.Sprintf("%v", mysql_driver_str)
	conf.MysqlDb.Driver = mysql_driver
	mysql_host_str := viper.Get("mysql.host")
	mysql_host := fmt.Sprintf("%v", mysql_host_str)
	conf.MysqlDb.Host = mysql_host
	mysql_port_str := viper.Get("mysql.port")
	mysql_port := fmt.Sprintf("%v", mysql_port_str)
	conf.MysqlDb.Port = mysql_port
	user_str := viper.Get("mysql.user")
	user_name := fmt.Sprintf("%v", user_str)
	conf.MysqlDb.User = user_name
	passwd_str := viper.Get("mysql.passwd")
	passwd := fmt.Sprintf("%v", passwd_str)
	conf.MysqlDb.Passwd = passwd
	// mysql database
	database_str := viper.Get("mysql.database")
	database_name := fmt.Sprintf("%v", database_str)
	conf.MysqlDb.Database = database_name
	// mysql database contains tables
	mysql_tables_str := fmt.Sprintf("%v", viper.Get("mysql.tables"))
	conf.MysqlDb.TableNames = strings.Split(strings.Trim(mysql_tables_str, "[]"), " ")
	mysql_alarm_list_str := fmt.Sprintf("%v", viper.Get("mysql.alarm_list"))
	conf.MysqlDb.AlarmList = strings.Split(strings.Trim(mysql_alarm_list_str, "[]"), " ")
	// monitor checkInterval
	conf.Monitor.CheckInterval = viper.GetInt("monitor.checkInterval")
	// monitor excludedServices
	monitor_excludedServices_str := fmt.Sprintf("%v", viper.Get("monitor.excludedServices"))
	conf.Monitor.ExcludedServices = strings.Split(strings.Trim(monitor_excludedServices_str, "[]"), " ")
	// model channel
	model_channel_list_str := fmt.Sprintf("%v", viper.Get("modelChannels"))
	conf.ModelChannels = strings.Split(strings.Trim(model_channel_list_str, "[]"), " ")
	// modelTransmit srcHost
	conf.ModelTransmit.SrcHost = viper.GetString("model_transmit.srcHost")
	// modelTransmit srcPath
	conf.ModelTransmit.SrcPath = viper.GetString("model_transmit.srcPath")
	// run (dev/prod)
	conf.RunEnv = viper.GetString("run_env")

	conf.LoadThreshold.CpuLimit = viper.GetFloat64("load_threshold.cpuLimit")
	conf.LoadThreshold.CheckInterval = viper.GetInt("load_threshold.checkinterval")
	conf.LoadThreshold.Method = viper.GetString("load_threshold.metheod")
	conf.LoadThreshold.Up_Gap = viper.GetInt("load_threshold.up_gap")
	conf.LoadThreshold.Down_Gap = viper.GetInt("load_threshold.down_gap")
	// elastic_expansion checkInterval
	conf.ElasticExpansion.CheckInterval = viper.GetInt("elastic_expansion.checkInterval")
	// service static list config
	conf.DataDir = viper.GetString("data_dir")
	conf.ServiceStaticList.UpdateInterval = viper.GetInt("service_static_list.updateInterval")
	conf.ServiceStaticList.AdServerRelayHost = viper.GetString("service_static_list.adServerRelayHost")
	conf.ServiceStaticList.PredictorRelayHost = viper.GetString("service_static_list.predictorRelayHost")
	conf.ServiceStaticList.SubDataDir = viper.GetString("service_static_list.subDataDir")
	// elastic_expansion service_to_host_nums
	var service_groups []ServiceGroup
	service_groups_i := viper.Get("elastic_expansion.service_groups").([]interface{})
	for _, service_group_map_i := range service_groups_i {
		service_group_map := service_group_map_i.(map[string]interface{})
		var service_group ServiceGroup
		services := service_group_map["services"].([]interface{})
		for _, service := range services {
			service_group.Services = append(service_group.Services, service.(string))
		}
		service_groups = append(service_groups, service_group)
	}
	conf.ElasticExpansion.ServiceGroups = service_groups
	// predictor_client
	conf.PredictorClient.ConsulAddress = viper.GetString("predictor_client.consul_address")
	conf.StressTestService = viper.GetString("stress_test_service")
	conf.ConsistenceService = viper.GetString("consistence_service")
	conf.Prometheus.Address = viper.GetString("prometheus.address")
	exclude_hosts_str := fmt.Sprintf("%v", viper.Get("migrate_hosts.exclude_hosts"))
	conf.MigrateHosts.ExcludeHosts = strings.Split(strings.Trim(exclude_hosts_str, "[]"), " ")
}

func GetConf() *Conf {
	return conf
}
