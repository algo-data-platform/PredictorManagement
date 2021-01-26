package env

import (
	"content_service/common"
	"content_service/conf"
	"content_service/db"
	"content_service/libs/logger"
	"encoding/json"
	"github.com/jinzhu/gorm"
)

type Env struct {
	Conf    *conf.Conf
	Db      *gorm.DB
	LocalIp string
}

// 初始化 zap logger
func InitLog(conf *conf.Conf) {
	logOptions := logger.New()
	logOptions.SetLogFile(conf.Log.FileName)
	logOptions.SetLevel(conf.Log.Level)
	logOptions.SetRotate(conf.Log.IsRotate)
	logOptions.SetRotateCycle(conf.Log.RotateCycle)
	logOptions.SetRotateMaxHours(conf.Log.RotateMaxHours)
	logOptions.InitLogger()

	conf_json, err := json.MarshalIndent(*conf, "", "  ")
	if err != nil {
		logger.Fatalf("invalid config! conf=%+v", *conf)
	} else {
		logger.Infof(`
                ///////////////////////////
                /////  Env Initiated  /////
                ///////////////////////////`)
		logger.Infof("loaded config: %v", string(conf_json))
	}
}

func New(conf *conf.Conf) *Env {
	db := db.New(conf)
	localIp, err := common.GetLocalIp(conf)
	if err != nil {
		logger.Fatalf("GetLocalIp() error, err: %v", err)
	}
	return &Env{
		Conf:    conf,
		Db:      db,
		LocalIp: localIp,
	}
}

func (env *Env) IsRegressionMode() bool {
	if common.IsInSliceString(env.LocalIp, env.Conf.RegressionService.Host) {
		return true
	}

	return false
}
