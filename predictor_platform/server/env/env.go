package env

import (
	"github.com/jinzhu/gorm"
	"io"
	"net"
	"server/conf"
	"server/libs/logger"
	"server/libs/mail"
	"server/libs/prometheus"
	"server/util"
)

type Environment struct {
	Conf      *conf.Conf
	LocalIp   string
	LogWriter io.Writer
	MysqlDB   *gorm.DB
	Mailer    *mail.Mail
	Prom      *prometheus.Prom
}

var Env *Environment

func New() {
	conf.New()
	local_ip, err := GetLocalIp(conf.GetConf())
	if err != nil {
		logger.Fatalf("GetLocalIp() error, err: %v", err)
	}
	logWriter := initLog(conf.GetConf())
	mysqlDB := InitMysql(conf.GetConf())
	mailer := mail.New(conf.GetConf().HostEmail, conf.GetConf().UsernameEmail,
		conf.GetConf().PasswdEmail)
	prom := prometheus.New("http://" + conf.GetConf().Prometheus.Address)
	Env = &Environment{
		Conf:      conf.GetConf(),
		LocalIp:   local_ip,
		LogWriter: logWriter,
		MysqlDB:   mysqlDB,
		Mailer:    mailer,
		Prom:      prom,
	}
}

func GetLocalIp(conf *conf.Conf) (string, error) {
	var ip = conf.HttpHost
	if ip != "LOCAL_IP" && ip != "" && net.ParseIP(ip) != nil {
		return ip, nil
	}
	return util.GetLocalIp()
}
