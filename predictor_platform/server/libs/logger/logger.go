package logger

import (
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"strings"
	"time"
)

const (
	_defaultEncoding = "console"
)

var (
	logger *zap.Logger
	// sugar logger
	sugarLogger               *zap.SugaredLogger
	_encoderNameToConstructor = map[string]func(zapcore.EncoderConfig) zapcore.Encoder{
		"console": func(encoderConfig zapcore.EncoderConfig) zapcore.Encoder {
			return zapcore.NewConsoleEncoder(encoderConfig)
		},
		"json": func(encoderConfig zapcore.EncoderConfig) zapcore.Encoder {
			return zapcore.NewJSONEncoder(encoderConfig)
		},
	}
	writeSyncer io.Writer
)

// 日志级别 debug < info < warn < error < panic < fatal
var levelMap = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
	"panic": zapcore.PanicLevel,
	"fatal": zapcore.FatalLevel,
}

type LogOptions struct {
	Encoding       string // "console" or "json"
	LogFileName    string // log file name
	Level          string // debug info warn error panic
	IsRotate       bool   // is rotate file
	RotateCycle    string // 日志切割周期 day or hour or minute
	RotateMaxHours int    // 保存最大时长，单位 小时
}

func init() {
	encoderConfig := getEncoderConfig()
	// 设置日志级别
	atom := zap.NewAtomicLevelAt(zap.InfoLevel)
	config := zap.Config{
		Level:            atom, // 日志级别
		Development:      false,
		Encoding:         "console",          // 输出格式 console 或 json
		EncoderConfig:    encoderConfig,      // 编码器配置
		OutputPaths:      []string{"stdout"}, // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
		ErrorOutputPaths: []string{"stderr"},
	}

	// 构建日志
	logger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("log init fail: %v", err))
	}
	sugarLogger = logger.Sugar()
}

func New() *LogOptions {
	return &LogOptions{
		Encoding: _defaultEncoding,
	}
}

func (c *LogOptions) SetEncoding(encoding string) {
	c.Encoding = encoding
}

func (c *LogOptions) SetLogFile(path string) {
	c.LogFileName = path
}

//日志级别 debug < info < warn < error < panic < fatal
func (c *LogOptions) SetLevel(level string) {
	c.Level = level
}

func (c *LogOptions) SetRotate(isRotate bool) {
	c.IsRotate = isRotate
}

func (c *LogOptions) SetRotateMaxHours(rotateMaxHours int) {
	c.RotateMaxHours = rotateMaxHours
}

func (c *LogOptions) SetRotateCycle(rotateCycle string) {
	c.RotateCycle = rotateCycle
}

// 初始化 logger 和 sugarLogger
func (c *LogOptions) InitLogger() {
	level := getLoggerLevel(c.Level)
	if c.IsRotate {
		writeSyncer = c.getRotateWriter()
	} else {
		writeSyncer = c.getDefaultWriter()
	}
	if c.Encoding == "" {
		c.Encoding = _defaultEncoding
	}
	encoder := _encoderNameToConstructor[c.Encoding]
	encoderConfig := getEncoderConfig()

	var core zapcore.Core
	core = zapcore.NewCore(encoder(encoderConfig), zapcore.AddSync(writeSyncer), level)

	stackTrace := zap.AddStacktrace(zapcore.PanicLevel)
	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), stackTrace)

	sugarLogger = logger.Sugar()
}

func getLoggerLevel(lvl string) zapcore.Level {
	if level, ok := levelMap[lvl]; ok {
		return level
	}
	return zapcore.InfoLevel
}

func (c *LogOptions) GetWriter() io.Writer {
	return writeSyncer
}

func (c *LogOptions) getRotateWriter() io.Writer {
	fileName := c.LogFileName
	// 生成rotatelogs的Logger 实际生成的文件名 content_service.log.YYmmddHH
	// 保存30天内的日志，每天(整点)分割一次日志
	hook, err := rotatelogs.New(
		strings.Join([]string{fileName, getRotateCycleFormat(c.RotateCycle)}, "."), // Y 年 m 月 d 日 H 时 M 分
		rotatelogs.WithLinkName(strings.Join([]string{fileName, "link"}, "-")),     // 会创建一个软链指向对应的最新切割日志
		rotatelogs.WithMaxAge(getRotateMaxAge(c.RotateMaxHours)),
		rotatelogs.WithRotationTime(getRotateTime(c.RotateCycle)),
	)

	if err != nil {
		panic(fmt.Sprintf("fail to rotate file, err: %v", err))
	}
	return hook
}

func (c *LogOptions) getDefaultWriter() io.Writer {
	fileName := c.LogFileName
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("fail open log file, err: %v", err))
	}
	return f
}

// 获取切割的周期可读格式
func getRotateCycleFormat(rotateCycle string) string {
	var cycleFormat string = "%Y%m%d"
	switch rotateCycle {
	case "day":
		cycleFormat = "%Y%m%d"
	case "hour":
		cycleFormat = "%Y%m%d%H"
	case "minute":
		cycleFormat = "%Y%m%d%H%M"
	}
	return cycleFormat
}

// 获取切割日志频率
func getRotateTime(rotateCycle string) time.Duration {
	var rotateTime time.Duration = time.Hour * 24
	switch rotateCycle {
	case "day":
		rotateTime = time.Hour * 24
	case "hour":
		rotateTime = time.Hour
	case "minute":
		rotateTime = time.Minute
	}
	return rotateTime
}

// 获取切割日志的最大保留时长
func getRotateMaxAge(rotateMaxHours int) time.Duration {
	var rotateMaxAge time.Duration = time.Hour * 24 * 7
	if rotateMaxHours > 0 {
		rotateMaxAge = time.Hour * time.Duration(rotateMaxHours)
	}
	return rotateMaxAge
}

func getEncoderConfig() zapcore.EncoderConfig {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "file",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     formatEncodeTime,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	return encoderConfig
}

func formatEncodeTime(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	// 这种格式会比下面的方式性能要好
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func Debug(msg string, args ...zap.Field) {
	logger.Debug(msg, args...)
}

func Info(msg string, args ...zap.Field) {
	logger.Info(msg, args...)
}

func Warn(msg string, args ...zap.Field) {
	logger.Warn(msg, args...)
}

func Error(msg string, args ...zap.Field) {
	logger.Error(msg, args...)
}

// panic to stdout
func Panic(msg string, args ...zap.Field) {
	logger.Panic(msg, args...)
}

// print msg ,then exit
func Fatal(msg string, args ...zap.Field) {
	logger.Fatal(msg, args...)
}

func Debugf(template string, args ...interface{}) {
	sugarLogger.Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	sugarLogger.Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	sugarLogger.Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	sugarLogger.Errorf(template, args...)
}

func DPanicf(template string, args ...interface{}) {
	sugarLogger.DPanicf(template, args...)
}

// panic to stdout
func Panicf(template string, args ...interface{}) {
	sugarLogger.Panicf(template, args...)
}

// print msg ,then exit
func Fatalf(template string, args ...interface{}) {
	sugarLogger.Fatalf(template, args...)
}
