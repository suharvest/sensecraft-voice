package recording

import (
	"github.com/gin-gonic/gin"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/middleware"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/options"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller"
)

type recordingRouter struct {
	c controller.SensecraftVoiceInterface
}

func NewRouter(o *options.Options) {
	router := &recordingRouter{c: o.Controller}
	router.initRoutes(o, o.HttpEngine)
}

func (r *recordingRouter) initRoutes(o *options.Options, httpEngine *gin.Engine) {
	// 设备侧路由按单条挂设备鉴权（WS 在握手阶段校验 header）
	deviceAuth := middleware.DeviceAuth(o)

	group := httpEngine.Group("/api/v1/recordings")
	{
		group.GET("/stream", deviceAuth, r.wsStream)
		group.POST("", deviceAuth, r.save)            // 新增HTTP POST接口
		group.POST("/batch", deviceAuth, r.saveBatch) // 客户端离线缓存批量上报
		group.GET("", r.list)
		group.GET("/keyword-matches", r.getKeywordMatches) // 新增关键词匹配查询接口
	}

	// 单独的查询接口，使用不同的路径keyword-matches
	queryGroup := httpEngine.Group("/api/recordings")
	{
		queryGroup.POST("/query", r.query)
	}
}
