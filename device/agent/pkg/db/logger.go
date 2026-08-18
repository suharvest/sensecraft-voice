package db

import (
	"context"
	"time"

	"gorm.io/gorm/logger"
)

type (
	SQLs []string

	DBLogger struct {
		logger.LogLevel
		SlowThreshold time.Duration // slow SQL queries
	}
)

const SQLContextKey = "sqls"

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
	if v := ctx.Value(SQLContextKey); v != nil {
		sqls := v.(*SQLs)
		*sqls = append(*sqls, sql)
	}
}

func WithDBContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, SQLContextKey, new(SQLs))
}

// GetSQLs returns all the SQL statements executed in the current context.
func GetSQLs(ctx context.Context) SQLs {
	if v := ctx.Value(SQLContextKey); v != nil {
		return *v.(*SQLs)
	}
	return SQLs{}
}
