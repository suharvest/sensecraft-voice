package middleware

import (
	"fmt"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"

	logutil "github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/log"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/sqlcontext"
)

func Logger(cfg *logutil.LogOptions) gin.HandlerFunc {
	return func(c *gin.Context) {
		l := logutil.NewLogger(cfg)
		c.Set(sqlcontext.SQLContextKey, new(sqlcontext.SQLs)) // set SQL context key

		// 处理请求操作
		c.Next()

		var err error
		if errs := c.Errors; len(errs) > 0 {
			err = fmt.Errorf("%v", errs.Errors())
		}
		l.WithLogFields(map[string]interface{}{
			"request_id":  requestid.Get(c),
			"method":      c.Request.Method,
			"uri":         c.Request.RequestURI,
			"status_code": c.Writer.Status(),
			"client_ip":   c.ClientIP(),
		})
		l.Log(c, err)
	}
}
