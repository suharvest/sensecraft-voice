package location

import (
	"github.com/gin-gonic/gin"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/options"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller"
)

type locationRouter struct {
	c controller.SensecraftVoiceInterface
}

func NewRouter(o *options.Options) {
	router := &locationRouter{c: o.Controller}
	router.initRoutes(o.HttpEngine)
}

func (r *locationRouter) initRoutes(httpEngine *gin.Engine) {
	group := httpEngine.Group("/api/v1/locations")
	{
		group.POST("", r.create)
		group.GET("", r.list)
		group.GET("/:id", r.getById)
		group.PUT("/:id", r.update)
		group.DELETE("/:id", r.delete)
	}

	// 按门店查询点位的路由
	storesGroup := httpEngine.Group("/api/v1/stores/:id/locations")
	{
		storesGroup.GET("", r.listByStoreId)
	}
}
