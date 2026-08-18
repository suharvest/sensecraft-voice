package db

import (
	"context"
	"time"

	"gorm.io/gorm/logger"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/sqlcontext"
)

type (
	DBLogger struct {
		logger.LogLevel
		SlowThreshold time.Duration // slow SQL queries
	}
)

func NewLogger(level logger.LogLevel, slowThreshold time.Duration) *DBLogger {
	return &DBLogger{
		LogLevel:      level,
		SlowThreshold: slowThreshold,
	}
}

func (l *DBLogger) LogMode(level logger.LogLevel) logger.Interface {
	l.LogLevel = level
	return l
}

func (l *DBLogger) Info(ctx context.Context, msg string, data ...interface{}) {}

func (l *DBLogger) Warn(ctx context.Context, msg string, data ...interface{}) {}

func (l *DBLogger) Error(ctx context.Context, msg string, data ...interface{}) {}

func (l *DBLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	sql, _ := fc()
	sqlcontext.AddSQL(ctx, sql)
}

func WithDBContext(ctx context.Context) context.Context {
	return sqlcontext.WithSQLContext(ctx)
}

// GetSQLs returns all the SQL statements executed in the current context.
func GetSQLs(ctx context.Context) sqlcontext.SQLs {
	return sqlcontext.GetSQLs(ctx)
}
