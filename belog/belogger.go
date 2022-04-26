package belog

import (
	"io"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// use zap Implement krakend/logging interface
type BELogger struct {
	filePath     string
	maxSize      int
	logger       *zap.Logger
	formatLogger *zap.SugaredLogger
	cfg          *zap.Config
	logLevel     zapcore.Level
}

var (
	instance *BELogger
	once     sync.Once
)

func InitZapLogger(filePath string, maxSize int, level zapcore.Level) (*BELogger, error) {
	belogger := &BELogger{
		filePath: filePath,
		maxSize:  maxSize,
		logLevel: level,
	}
	belogger.Init(0, nil)

	return belogger, nil
}

func NewZapLogger(filePath string, maxSize int, level zapcore.Level, ws ...io.Writer) (*BELogger, error) {
	belogger := &BELogger{
		filePath: filePath,
		maxSize:  maxSize,
		logLevel: level,
	}

	return belogger, nil
}

func GetLogger() *BELogger {
	once.Do(func() {
		instance = &BELogger{
			filePath: "./belog.txt",
			maxSize:  500,
			logLevel: zapcore.DebugLevel,
		}
	})

	return instance
}

func (l *BELogger) Init(callerSkipVal int, ws ...io.Writer) {
	l.cfg = l.newConfig()

	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   l.filePath,
		MaxSize:    l.maxSize, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})

	// TODO: Write new krakend-gologging and gelf's log format encoder, it is
	// similar zapcore.JSONEncoder
	//var writeSyncer []zapcore.WriteSyncer
	//if ws != nil {
	//for _, e := range ws {
	//writeSyncer = append(writeSyncer, zapcore.AddSync(e))
	//}
	//}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(l.cfg.EncoderConfig),
		w,
		l.logLevel,
	)

	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(callerSkipVal))

	zap.ReplaceGlobals(logger)
	l.logger = zap.L()
	l.formatLogger = zap.S()
}

func (l *BELogger) FlushBuffer() {
	l.logger.Sync()
}

func (l *BELogger) newConfig() *zap.Config {
	return &zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Development:      false,
		Encoding:         "json",
		OutputPaths:      []string{"stdout", l.filePath},
		ErrorOutputPaths: []string{"stderr", l.filePath},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}
}

func (l *BELogger) SetLogLevel(level zapcore.Level) {
	l.logLevel = level
}

func (l *BELogger) SetFilePath(filePath string) {
	l.filePath = filePath
}

func (l *BELogger) SetMaxSize(maxSize int) {
	l.maxSize = maxSize
}

func (l *BELogger) Debug(args ...interface{}) {
	l.formatLogger.Debug(args)
}

func (l *BELogger) Debugf(template string, args ...interface{}) {
	l.formatLogger.Debug(args)
}

func (l *BELogger) Debugw(msg string, keysAndValues ...interface{}) {
	l.formatLogger.Debugw(msg, keysAndValues)
}

func (l *BELogger) Info(args ...interface{}) {
	l.formatLogger.Info(args)
}

func (l *BELogger) Infof(template string, args ...interface{}) {
	l.formatLogger.Infof(template, args)
}

func (l *BELogger) Infow(msg string, keysAndValues ...interface{}) {
	l.formatLogger.Infow(msg, keysAndValues)
}

func (l *BELogger) Warn(args ...interface{}) {
	l.formatLogger.Warn(args)
}

func (l *BELogger) Warning(args ...interface{}) {
	l.formatLogger.Warn(args)
}

func (l *BELogger) Warnf(template string, args ...interface{}) {
	l.formatLogger.Warnf(template, args)
}

func (l *BELogger) Warnw(msg string, keysAndValues ...interface{}) {
	l.formatLogger.Warnw(msg, keysAndValues)
}

func (l *BELogger) Error(args ...interface{}) {
	l.formatLogger.Error(args)
}

func (l *BELogger) Errorf(template string, args ...interface{}) {
	l.formatLogger.Errorf(template, args)
}

func (l *BELogger) Errorw(msg string, keysAndValues ...interface{}) {
	l.formatLogger.Errorw(msg, keysAndValues)
}

func (l *BELogger) Critical(args ...interface{}) {
	l.formatLogger.Fatal(args)
}

func (l *BELogger) Fatal(args ...interface{}) {
	l.formatLogger.Fatal(args)
}

func (l *BELogger) Fatalf(template string, args ...interface{}) {
	l.formatLogger.Fatalf(template, args)
}

func (l *BELogger) Fatalw(msg string, keysAndValues ...interface{}) {
	l.formatLogger.Fatalw(msg, keysAndValues)
}
