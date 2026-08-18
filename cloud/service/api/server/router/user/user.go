package user

import (
	"github.com/gin-gonic/gin"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/options"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller"
)

type userRouter struct {
	c controller.SensecraftVoiceInterface
}

func NewRouter(o *options.Options) {
	router := &userRouter{c: o.Controller}
	router.initRoutes(o.HttpEngine)
}

func (r *userRouter) initRoutes(httpEngine *gin.Engine) {
	group := httpEngine.Group("/api/v1/users")
	{
		group.POST("/register", r.register)
		group.POST("/login", r.login)
		group.GET("", r.list)
		group.GET("/:id", r.getById)
		group.PUT("/:id", r.update)
		group.DELETE("/:id", r.delete)
		group.PUT("/:id/password", r.changePassword)
	}
}
