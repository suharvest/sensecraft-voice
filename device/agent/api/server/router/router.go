package router

import (
	"net/http"

	"github.com/YOUR-ORG/sensecraft-voice/device/agent/api/server/router/mqtt"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/api/server/router/voice"

	_ "github.com/YOUR-ORG/sensecraft-voice/device/agent/api/server/validator"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/YOUR-ORG/sensecraft-voice/device/agent/api/server/middleware"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/api/server/router/audit"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/cmd/app/options"
)

type RegisterFunc func(o *options.Options)

func InstallRouters(o *options.Options) {
	fs := []RegisterFunc{
		middleware.InstallMiddlewares,
		audit.NewRouter,
		mqtt.NewRouter,
		voice.NewRouter,
	}

	install(o, fs...)

	// 静态文件服务 - 提供前端页面
	o.HttpEngine.Static("/web", "./web")

	// 根路径重定向到首页
	o.HttpEngine.GET("/", func(c *gin.Context) {
		c.File("./web/index.html")
	})

	// 静态文件服务 - 提供前端页面
	// o.HttpEngine.Static("/web", "/app/web")

	// // 根路径重定向到首页
	// o.HttpEngine.GET("/", func(c *gin.Context) {
	// 	c.File("/app/web/index.html")
	// })

	// 启动健康检查
	o.HttpEngine.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	// 启动 APIs 服务
	o.HttpEngine.GET("/api-ref/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func install(o *options.Options, fs ...RegisterFunc) {
	for _, f := range fs {
		f(o)
	}
}
