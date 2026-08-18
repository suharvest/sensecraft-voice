package sqlcontext

import (
	"context"
)

type (
	SQLs []string
)

const SQLContextKey = "sqls"

// WithSQLContext 创建一个包含SQL上下文的context
func WithSQLContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, SQLContextKey, new(SQLs))
}

// GetSQLs 返回当前上下文中执行的所有SQL语句
func GetSQLs(ctx context.Context) SQLs {
	if v := ctx.Value(SQLContextKey); v != nil {
		return *v.(*SQLs)
	}
	return SQLs{}
}

// AddSQL 向上下文中添加SQL语句
func AddSQL(ctx context.Context, sql string) {
	if v := ctx.Value(SQLContextKey); v != nil {
		sqls := v.(*SQLs)
		*sqls = append(*sqls, sql)
	}
}
