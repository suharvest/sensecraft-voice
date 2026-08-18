package store

import (
	"github.com/gin-gonic/gin"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/options"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller"
)

type storeRouter struct {
	c controller.SensecraftVoiceInterface
}

func NewRouter(o *options.Options) {
	router := &storeRouter{c: o.Controller}
	router.initRoutes(o.HttpEngine)
}

func (r *storeRouter) initRoutes(httpEngine *gin.Engine) {
	group := httpEngine.Group("/api/v1/stores")
	{
		group.POST("", r.create)
		group.GET("", r.list)
		group.GET("/:id", r.getById)
		group.PUT("/:id", r.update)
		group.DELETE("/:id", r.delete)
	}
}
