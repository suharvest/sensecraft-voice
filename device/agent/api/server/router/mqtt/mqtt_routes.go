package mqtt

import (
	"github.com/gin-gonic/gin"

	"github.com/YOUR-ORG/sensecraft-voice/device/agent/api/server/httputils"
)

func (a *mqttRouter) push(c *gin.Context) {
	r := httputils.NewResponse()

	if err := a.c.Mqtt().Push(c); err != nil {
		httputils.SetFailed(c, r, err)
		return
	}
	httputils.SetSuccess(c, r)
}
