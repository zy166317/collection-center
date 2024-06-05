package logger

import (
	"fmt"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

var logger *zap.SugaredLogger

type LogConfig struct {
	LogPath string
}

func init() {
	if logger == nil {
		InitLog("")
	}
}

func InitLog(logPath string, logLevel ...string) {
	//var core zapcore.Core
	//var zapLevel zapcore.Level
	//zapLevel.Set(level)
	//
	//encoder := NewEncoderConfig()
	//if c.LogAddress == "" {
	//	encoder.EncodeTime = TimeEncoder
	//	syncWriter := zapcore.AddSync(os.Stdout)
	//	core = zapcore.NewCore(zapcore.NewJSONEncoder(encoder), syncWriter, zap.NewAtomicLevelAt(zapLevel))
	//	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	//	logger = zapLogger.Sugar()
	//} else {
	//	encoder.EncodeLevel = gelf.SyslogLevelEncoder
	//	encoder.TimeKey = "timestamp"
	//	encoder.EncodeTime = TimeEncoder
	//	syncWriter := gelf.New(c.LogAddress)
	//	core = zapcore.NewCore(zapcore.NewJSONEncoder(encoder), syncWriter, zap.NewAtomicLevelAt(zapLevel))
	//	fields := zap.Fields(zap.String("host", c.LogHost), zap.String("tag", c.LogTag))
	//	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), fields)
	//	logger = zapLogger.Sugar()
	//}
	// 设置一些基本日志格式 具体含义还比较好理解，直接看zap源码也不难懂
	encoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey:   "msg",
		LevelKey:     "level",
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		TimeKey:      "ts",
		EncodeTime:   TimeEncoder,
		CallerKey:    "file",
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(int64(d) / 1000000)
		},
	})
	// 实现两个判断日志等级的interface (其实 zapcore.*Level 自身就是 interface)
	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.WarnLevel
	})

	warnLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.WarnLevel
	})
	var ostype = runtime.GOOS
	var rootPath = ""
	if ostype == "windows" && logPath != "" && !strings.Contains(logPath, ":") {
		rootPath = "c:"
	}
	// 获取 info、warn日志文件的io.Writer 抽象 getWriter() 在下方实现
	infoWriter := getDateWriter(rootPath+logPath, "info.log")
	warnWriter := getDateWriter(rootPath+logPath, "error.log")
	// 最后创建具体的Logger
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(infoWriter), infoLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(warnWriter), warnLevel),
	)
	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
}
func GetLogger() *zap.SugaredLogger {
	return logger
}

func getWriter(dir string, filename string) io.Writer {
	if dir == "" {
		return os.Stdout
	}
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		fmt.Printf("%s", err)
	}
	hook := lumberjack.Logger{
		Filename:   path.Join(dir, filename), // 日志文件路径
		MaxSize:    128,                      // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: 100,                      // 日志文件最多保存多少个备份
		MaxAge:     30,                       // 文件最多保存多少天
		Compress:   true,                     // 是否压缩
	}
	if runtime.GOOS == "windows" {
		hook.Filename = windowsPathJoin(dir, filename)
	}
	// 生成rotatelogs的Logger 实际生成的文件名 demo.log.YYmmddHH
	// demo.log是指向最新日志的链接
	// 保存30天内的日志，每12小时(整点)分割一次日志
	//hook, err := rotatelogs.New(
	//	//	dir+filename+".%Y%m%d-%H", // 没有使用go风格反人类的format格式
	//	//	rotatelogs.WithMaxAge(time.Hour*24*30),
	//	//	rotatelogs.WithRotationTime(time.Hour*12),
	//	//)
	//	//if err != nil {
	//	//	panic(err)
	//	//}
	return io.Writer(&hook)
}

func getDateWriter(path string, filename string) io.Writer {
	if path == "" {
		return os.Stdout
	}
	// 生成rotatelogs的Logger 实际生成的文件名 YYmmddHH.demo.log
	// demo.log是指向最新日志的链接
	// 保存30天内的日志，每12小时(整点)分割一次日志
	err := os.MkdirAll(path, 0777)
	if err != nil {
		fmt.Printf("%s", err)
	}
	hook, err := rotatelogs.New(
		path+"%Y%m%d."+filename,
		rotatelogs.WithMaxAge(time.Hour*24*30),
		rotatelogs.WithRotationTime(time.Hour*24),
	)
	if err != nil {
		panic(err)
	}
	return hook
}

func windowsPathJoin(elem ...string) string {
	for i, e := range elem {
		if e != "" {
			return strings.Join(elem[i:], "\\")
		}
	}
	return ""
}
func NewEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "name",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "trace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	logger.Debugf(template, args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Infof(template string, args ...interface{}) {
	logger.Infof(template, args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	logger.Warnf(template, args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	logger.Errorf(template, args...)
}

func Panic(args ...interface{}) {
	logger.Panic(args...)
}

func Panicf(template string, args ...interface{}) {
	logger.Panicf(template, args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	logger.Fatalf(template, args...)
}
