package audit

import (
	"github.com/gin-gonic/gin"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/options"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller"
)

type auditRouter struct {
	c controller.SensecraftVoiceInterface
}

func NewRouter(o *options.Options) {
	router := &auditRouter{
		c: o.Controller,
	}
	router.initRoutes(o.HttpEngine)
}

func (a *auditRouter) initRoutes(httpEngine *gin.Engine) {
	auditRoute := httpEngine.Group("/api/v1/audits")
	{
		// get 日志
		auditRoute.GET("/:auditId", a.getAudit)
		auditRoute.GET("", a.listAudits)
	}
}
