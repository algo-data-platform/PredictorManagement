package conf

import (
	"content_service/libs/logger"
	"encoding/json"
	"flag"
	"os"
	"time"
)

type Conf struct {
	LocalIp              string
	Stage                string
	HttpPort             string
	DingDingWebhookUrl   string
	EnalbleRouterService bool
	Log                  struct {
		FileName       string
		Level          string
		IsRotate       bool
		RotateCycle    string
		RotateMaxHours int
	}
	Db struct {
		Driver string
		Host   string
		Port   string
		User   string
		Passwd string
		Name   string
	}
	Services        []string
	P2PModelService struct {
		RunInterval   time.Duration
		TargetService struct {
			HttpPort string
		}
		SrcHost               string
		SrcPath               string
		DestHost              string
		DestPath              string
		Retry                 int
		RsyncBWLimit          int
		SrcRsyncBWLimit       int
		PeerLimit             int
		ServicePullAlertLimit int
		ModelPullMaxLimit     int
	}
	CleaningService struct {
		RunInterval               time.Duration
		VersionsToKeep            int
		VersionsToKeepForValidate int
		PathOverride              string
		HoursToKeep               int
	}
	ValidateService struct {
		Host             string
		RetryInterval    time.Duration
		RetryTimes       int
		MaxSampleCount   int
		PredictorTimeout int
		ConsulAddress    string
		ServiceName      string
		HtmlTemplateDir  string
		SummaryResultDir string
		AlgoLogDir       string
		AlgoLogBaseUrl   string
		ReportRecipients []string
	}
	HdfsService struct {
		RunInterval  time.Duration
		Host         string
		RsyncBWLimit int
		DestPath     string
		TransmitHost string
		TransmitPath string
	}
	FileSyncService struct {
		RunInterval            time.Duration
		RetryTimes             int
		RsyncBWLimit           int
		SrcRsyncBWLimit        int
		SrcHost                string
		SrcPath                string
		DestPath               string
		SyncTimesLimit         int
		PredictorStaticListDir string
	}
	Alert struct {
		Rate       int
		Recipients []string
	}
	StressTestService struct {
		ServiceName string
	}
	RegressionService struct {
		Host       []string
		ScriptPath string
		PacketName string
		SleepTime  int
	}
}

func New() *Conf {
	var conf_file string
	flag.StringVar(&conf_file, "conf", "", "config file path")
	flag.Parse()
	file, err := os.Open(conf_file)
	defer file.Close()
	if err != nil {
		logger.Panicf("open config file=%v failed!", conf_file)
	}
	decoder := json.NewDecoder(file)
	conf := &Conf{}
	err = decoder.Decode(conf)
	if err != nil {
		logger.Panicf("parse config file=%v failed! err: %v", conf_file, err)
	}
	return conf
}
