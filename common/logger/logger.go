package logger

import (
	"io"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

// error logger
var errorLogger *zap.SugaredLogger

var levelMap = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
	"fatal":  zapcore.FatalLevel,
}

func getLoggerLevel(lvl string) zapcore.Level {
	if level, ok := levelMap[lvl]; ok {
		return level
	}
	return zapcore.InfoLevel
}

func getWriter(filepath string) io.Writer {
	// 保存7天内的日志，每1小时(整点)分割一次日志
	hook, err := rotatelogs.New(
		filepath+"%Y-%m-%d.log",
		rotatelogs.WithLinkName(filepath+"compiler_service.log"),
		rotatelogs.WithMaxAge(time.Hour*24*28),
		rotatelogs.WithRotationTime(time.Hour*24*7),
	)
	if err != nil {
		panic(err)
	}
	return hook
}

func InitLogger(filePath, model string, out bool) {
	var logger *zap.Logger
	// default debug model
	level := getLoggerLevel(model)
	// file out
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		_ = os.Mkdir(filePath, os.ModePerm)
	}
	syncWriter := zapcore.AddSync(getWriter(filePath))
	//syncWriter := zapcore.AddSync(getWriter(fileName))
	encoder := zap.NewProductionEncoderConfig()
	encoder.EncodeTime = zapcore.ISO8601TimeEncoder
	//zapcore.NewConsoleEncoder(encoder)
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoder), syncWriter, zap.NewAtomicLevelAt(level))
	if out {
		outSyncWriter := zapcore.AddSync(os.Stdout)
		stdCore := zapcore.NewCore(zapcore.NewConsoleEncoder(encoder), outSyncWriter, zap.NewAtomicLevelAt(level))
		logger = zap.New(zapcore.NewTee(core, stdCore), zap.AddCaller(), zap.AddCallerSkip(1))
	} else {
		logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	}
	errorLogger = logger.Sugar()
	return
}

func Debug(args ...interface{}) {
	errorLogger.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	errorLogger.Debugf(template, args...)
}

func Info(args ...interface{}) {
	errorLogger.Info(args...)
}

func Infof(template string, args ...interface{}) {
	errorLogger.Infof(template, args...)
}

func Warn(args ...interface{}) {
	errorLogger.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	errorLogger.Warnf(template, args...)
}

func Error(args ...interface{}) {
	errorLogger.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	errorLogger.Errorf(template, args...)
}

func DPanic(args ...interface{}) {
	errorLogger.DPanic(args...)
}

func DPanicf(template string, args ...interface{}) {
	errorLogger.DPanicf(template, args...)
}

func Panic(args ...interface{}) {
	errorLogger.Panic(args...)
}

func Panicf(template string, args ...interface{}) {
	errorLogger.Panicf(template, args...)
}

func Fatal(args ...interface{}) {
	errorLogger.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	errorLogger.Fatalf(template, args...)
}
