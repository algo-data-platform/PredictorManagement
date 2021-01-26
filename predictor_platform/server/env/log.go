package env

import (
	"io"
	"server/conf"
	"server/libs/logger"
)

func initLog(conf *conf.Conf) io.Writer {
	logOptions := logger.New()
	logOptions.SetLogFile(conf.Log.FileName)
	logOptions.SetLevel(conf.Log.Level)
	logOptions.SetRotate(conf.Log.IsRotate)
	logOptions.SetRotateCycle(conf.Log.RotateCycle)
	logOptions.SetRotateMaxHours(conf.Log.RotateMaxHours)
	logOptions.InitLogger()
	return logOptions.GetWriter()
}
