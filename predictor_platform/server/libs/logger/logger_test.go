package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"testing"
	"time"
)

func BenchmarkZap(b *testing.B) {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "file",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	// 设置日志级别
	atom := zap.NewAtomicLevelAt(zap.InfoLevel)
	config := zap.Config{
		Level:            atom, // 日志级别
		Development:      false,
		Encoding:         "console",                          // 输出格式 console 或 json
		EncoderConfig:    encoderConfig,                      // 编码器配置
		OutputPaths:      []string{"/tmp/benchmark-zap.log"}, // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
		ErrorOutputPaths: []string{"stderr"},
	}

	// 构建日志
	logger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("log 初始化失败: %v", err))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("this is zap test",
			zap.String("url", "github.com"),
			zap.Int("num", 3),
			zap.Duration("second", time.Second),
		)
	}
}

func BenchmarkLogger(b *testing.B) {
	c := New()
	c.SetLogFile("/tmp/benchmark-logger.log")
	c.SetEncoding("console")
	c.SetLevel("debug")
	c.SetRotate(false)
	c.InitLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("this is logger test",
			zap.String("url", "github.com"),
			zap.Int("num", 3),
			zap.Duration("second", time.Second),
		)
	}
}

// 并发性能测试
func BenchmarkLoggerParallel(b *testing.B) {
	c := New()
	c.SetLogFile("/tmp/benchmark-zap-parallel.log")
	c.SetEncoding("console")
	c.SetLevel("debug")
	c.SetRotate(false)
	c.InitLogger()

	// 测试一个对象或者函数在多线程的场景下面是否安全
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Info("this is logger test parallel",
				zap.String("url", "github.com"),
				zap.Int("num", 3),
				zap.Duration("second", time.Second),
			)
		}
	})
}

func BenchmarkSuggarZap(b *testing.B) {
	c := New()
	c.SetLogFile("/tmp/benchmark-suggarzap.log")
	c.SetLevel("debug")
	c.SetEncoding("console")
	c.SetRotate(false)
	c.InitLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Infof("this is zap test,url:%s,num:%d,second:%v",
			"github.com",
			3,
			time.Second,
		)
	}
}

func BenchmarkLog(b *testing.B) {
	var full_file_name string
	full_file_name = "/tmp/benchmark-log.log"

	f, err := os.OpenFile(full_file_name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(f)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Printf("this is zap test,url:%s,num:%d,second:%v",
			"github.com",
			3,
			time.Second,
		)
	}
}

func BenchmarkLogrus(b *testing.B) {
	var log = logrus.New()
	//用日志实例的方式使用日志
	fileName := "/tmp/benchmark-logrus.log"
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("fail open log file, err: %v", err))
	}
	log.Out = f
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.WithFields(logrus.Fields{
			"url": "github.com",
			"num": 3,
		}).Info("this is logrus test")
	}
}
