package stats

import (
	"github.com/gin-gonic/gin"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/options"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller"
)

type statsRouter struct {
	c controller.SensecraftVoiceInterface
}

func NewRouter(o *options.Options) {
	router := &statsRouter{c: o.Controller}
	router.initRoutes(o.HttpEngine)
}

func (r *statsRouter) initRoutes(httpEngine *gin.Engine) {
	group := httpEngine.Group("/api/v1/stats")
	{
		group.GET("/dashboard", r.getDashboardStats)
	}
}
