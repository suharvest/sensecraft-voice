package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/options"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util"
)

var alwaysAllowPath sets.String

func init() {
	alwaysAllowPath = sets.NewString("/api/v1/users/login")
}

// 允许特定请求不经过验证
func allowCustomRequest(c *gin.Context) bool {
	// 用户请求
	if strings.HasPrefix(c.Request.URL.Path, "/api/v1/users") {
		switch c.Request.Method {
		case http.MethodPost:
			return c.Query("initAdmin") == "true"
		case http.MethodGet:
			return c.Query("count") == "true"
		}
	}

	// TODO: 其他请求
	return false
}

func InstallMiddlewares(o *options.Options) {
	// 依次进行跨域，日志，单用户限速，总量限速，验证，鉴权和审计
	o.HttpEngine.Use(
		requestid.New(requestid.WithGenerator(func() string {
			return util.GenerateRequestID()
		})),
		Cors(),
		Logger(&o.ComponentConfig.Default.LogOptions),
		Admission())
}
