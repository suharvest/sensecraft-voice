package chat

import (
	"github.com/gin-gonic/gin"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/options"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller"
)

type chatRouter struct {
	c controller.SensecraftVoiceInterface
}

func NewRouter(o *options.Options) {
	router := &chatRouter{c: o.Controller}
	router.initRoutes(o.HttpEngine)
}

func (r *chatRouter) initRoutes(httpEngine *gin.Engine) {
	group := httpEngine.Group("/api/v1/chat")
	{
		group.POST("/stream", r.streamMessage) // 流式响应接口
		group.POST("/send", r.sendMessage)     // 保留同步接口
		group.GET("/history/:session_id", r.getChatHistory)
		group.GET("/session/:session_id", r.getChatSession)
		group.GET("/sessions", r.getChatSessions)                         // 获取会话列表接口
		group.DELETE("/session/:session_id", r.deleteChatSession)         // 删除单个会话接口
		group.DELETE("/sessions", r.deleteChatSessions)                   // 批量删除会话接口
		group.PUT("/session/:session_id/title", r.updateChatSessionTitle) // 更新会话标题接口
	}
}
