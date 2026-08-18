package log

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	klog "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/db"
)

var (
	once sync.Once
	log  = klog.New()
)

// 导出常用的日志方法
var (
	Debug      = log.Debug
	Debugf     = log.Debugf
	Info       = log.Info
	Infof      = log.Infof
	Warn       = log.Warn
	Warnf      = log.Warnf
	Error      = log.Error
	Errorf     = log.Errorf
	Fatal      = log.Fatal
	Fatalf     = log.Fatalf
	WithField  = log.WithField
	WithFields = log.WithFields
)

type LogFormat string

const (
	LogFormatJson LogFormat = "json"
	LogFormatText LogFormat = "text"
)

var ErrInvalidLogFormat = errors.New("invalid log format")

type LogFileConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Path       string `yaml:"path"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

type LogOptions struct {
	LogFormat `yaml:"log_format"`
	LogSQL    bool          `yaml:"log_sql"`
	LogFile   LogFileConfig `yaml:"log_file"`
}

// DefaultLogOptions returns the default configs.
func DefaultLogOptions() *LogOptions {
	return &LogOptions{
		LogFormat: LogFormatJson,
		LogSQL:    false,
		LogFile: LogFileConfig{
			Enabled:    false,
			Path:       "./logs/app.log",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
		},
	}
}

func (o *LogOptions) Valid() error {
	switch o.LogFormat {
	case LogFormatJson, LogFormatText:
		return nil
	default:
		return ErrInvalidLogFormat
	}
}

// Init sets the log format only once.
func (o *LogOptions) Init() {
	once.Do(func() {
		// 设置日志输出到文件
		if o.LogFile.Enabled {
			writer := &lumberjack.Logger{
				Filename:   o.LogFile.Path,
				MaxSize:    o.LogFile.MaxSize,
				MaxBackups: o.LogFile.MaxBackups,
				MaxAge:     o.LogFile.MaxAge,
				Compress:   o.LogFile.Compress,
			}
			log.SetOutput(writer)
		}

		// 设置日志格式
		switch o.LogFormat {
		case LogFormatJson:
			log.SetFormatter(&klog.JSONFormatter{
				TimestampFormat: time.RFC3339Nano,
			})
		default:
			log.SetFormatter(&klog.TextFormatter{
				FullTimestamp:          true,
				TimestampFormat:        "2006-01-02 15:04:05",
				ForceColors:            true,
				DisableColors:          false,
				DisableLevelTruncation: true,
				PadLevelText:           true,
				DisableQuote:           true,
				DisableTimestamp:       false,
				DisableSorting:         true,
				QuoteEmptyFields:       true,
				FieldMap: klog.FieldMap{
					klog.FieldKeyTime:  "@timestamp",
					klog.FieldKeyLevel: "@level",
					klog.FieldKeyMsg:   "@message",
				},
			})
		}

		// 设置日志级别
		log.SetLevel(klog.InfoLevel)
	})
}

const (
	SuccessMsg = "SUCCESS"
	ErrorMsg   = "ERROR"
	FailMsg    = "FAIL"
)

type Logger struct {
	startTime time.Time
	logSQL    bool
	logEntry  *klog.Entry
}

func NewLogger(cfg *LogOptions) *Logger {
	return &Logger{
		startTime: time.Now(),
		logSQL:    cfg.LogSQL,
		logEntry:  klog.NewEntry(klog.StandardLogger()),
	}
}

func (l *Logger) WithLogField(key string, value interface{}) {
	l.logEntry = l.logEntry.WithField(key, value)
}

func (l *Logger) WithLogFields(fields map[string]interface{}) {
	l.logEntry = l.logEntry.WithFields(fields)
}

func (l *Logger) Log(ctx context.Context, err error) {
	fields := make(map[string]interface{})
	if l.logSQL {
		if sqls := db.GetSQLs(ctx); len(sqls) > 0 {
			fields["sqls"] = sqls
		}
	}
	fields["latency"] = fmt.Sprintf("%dµs", time.Since(l.startTime).Microseconds())

	if err != nil {
		fields["error"] = err
		l.logEntry.WithFields(fields).Error(FailMsg)
		return
	}

	l.logEntry.WithFields(fields).Info(SuccessMsg)
}
