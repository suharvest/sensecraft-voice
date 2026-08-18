package mqtt

import (
	"github.com/gin-gonic/gin"

	"github.com/YOUR-ORG/sensecraft-voice/device/agent/cmd/app/options"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/controller"
)

type mqttRouter struct {
	c controller.SensecraftVoiceInterface
}

func NewRouter(o *options.Options) {
	router := &mqttRouter{
		c: o.Controller,
	}
	router.initRoutes(o.HttpEngine)
}

func (a *mqttRouter) initRoutes(httpEngine *gin.Engine) {
	mqttRoute := httpEngine.Group("/api/v1/mqtt")
	{
		// get 日志
		mqttRoute.GET("/push", a.push)
	}
}
