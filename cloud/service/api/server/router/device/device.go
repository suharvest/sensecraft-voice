package device

import (
	"github.com/gin-gonic/gin"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/middleware"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/options"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller"
)

type deviceRouter struct {
	c controller.SensecraftVoiceInterface
}

func NewRouter(o *options.Options) {
	router := &deviceRouter{c: o.Controller}
	router.initRoutes(o, o.HttpEngine)
}

func (r *deviceRouter) initRoutes(o *options.Options, httpEngine *gin.Engine) {
	// 设备侧与 web 管理端路由混在同一 group，中间件只能按单条路由挂
	deviceAuth := middleware.DeviceAuth(o)
	registerAuth := middleware.DeviceRegisterAuth(o)
	userAuth := middleware.UserAuth(o)

	group := httpEngine.Group("/api/v1/devices")
	{
		// 设备侧：注册 + 心跳（enrollment key 或 device token）
		group.POST("/register", registerAuth, r.register)
		// 设备侧：拉取 ASR 配置。路径 /api/v1/devices/me/asr-config 由 :mac="me" 匹配
		// （gin 的路由树不允许静态段与通配段在同层并存，故复用 :mac 占位）
		group.GET("/:mac/asr-config", deviceAuth, r.getAsrConfig)

		group.GET("", r.list)
		// :mac 段兼容数字 id：handler 内判别，避免与 :id 同层冲突
		group.GET(":mac", r.getByMacOrId)

		// 管理端：改动类接口需要登录
		group.PUT("/:id", userAuth, r.update)
		group.PUT("/:id/assign", userAuth, r.assignToLocation)
		group.PUT("/:id/name", userAuth, r.updateName)
		group.PUT("/:id/asr-server", userAuth, r.assignAsrServer)
	}

	// 按点位查询设备
	locationsGroup := httpEngine.Group("/api/v1/locations/:id/devices")
	{
		locationsGroup.GET("", r.listByLocation)
	}

	// 按门店查询设备
	storesGroup := httpEngine.Group("/api/v1/stores/:id/devices")
	{
		storesGroup.GET("", r.listByStore)
	}
}
