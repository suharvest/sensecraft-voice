package router

import (
	"net/http"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/router/mqtt"

	_ "github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/validator"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/middleware"
	asrserver "github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/router/asr_server"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/router/audio_recordings"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/router/audit"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/router/chat"
	devicer "github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/router/device"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/router/keywords"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/router/location"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/router/openai"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/router/oss"
	recrouter "github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/router/recording"
	statsrouter "github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/router/stats"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/router/store"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/router/user"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/options"
)

type RegisterFunc func(o *options.Options)

func InstallRouters(o *options.Options) {
	fs := []RegisterFunc{
		middleware.InstallMiddlewares,
		audit.NewRouter,
		mqtt.NewRouter,
		devicer.NewRouter,
		recrouter.NewRouter,
		statsrouter.NewRouter,
		store.NewRouter,
		location.NewRouter,
		user.NewRouter,
		chat.NewRouter,
		openai.NewOpenAIRouter,
		oss.NewOSSRouter,
		audio_recordings.NewRouter,
		keywords.NewRouter,
		asrserver.NewRouter,
	}

	install(o, fs...)

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
