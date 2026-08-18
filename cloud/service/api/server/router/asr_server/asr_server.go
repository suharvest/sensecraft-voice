package asr_server

import (
	"github.com/gin-gonic/gin"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/middleware"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/options"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller"
)

type asrServerRouter struct {
	c controller.SensecraftVoiceInterface
}

func NewRouter(o *options.Options) {
	router := &asrServerRouter{c: o.Controller}
	router.initRoutes(o, o.HttpEngine)
}

// initRoutes ASR 服务器纳管接口。表里存有各服务器的 api_key，全组必须挂管理端鉴权。
func (r *asrServerRouter) initRoutes(o *options.Options, httpEngine *gin.Engine) {
	group := httpEngine.Group("/api/v1/asr-servers", middleware.UserAuth(o))
	{
		group.POST("", r.create)
		group.GET("", r.list)
		group.GET("/:id", r.getById)
		group.PUT("/:id", r.update)
		group.DELETE("/:id", r.delete)
		group.POST("/:id/probe", r.probe)
	}
}
