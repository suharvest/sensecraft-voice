package openai

import (
	"github.com/gin-gonic/gin"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/options"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller"
)

// openaiRouter OpenAI路由
type openaiRouter struct {
	c controller.SensecraftVoiceInterface
}

// NewOpenAIRouter 创建OpenAI路由
func NewOpenAIRouter(o *options.Options) {
	router := &openaiRouter{c: o.Controller}
	router.initRoutes(o.HttpEngine)
}

// initRoutes 注册OpenAI路由
func (r *openaiRouter) initRoutes(httpEngine *gin.Engine) {
	// v2版本的OpenAI接口
	group := httpEngine.Group("/api/v2/openai")

	// 聊天相关接口
	group.POST("/chat/send", r.sendMessage)
	group.POST("/chat/stream", r.streamMessage)
	group.GET("/chat/history/:session_id", r.getChatHistory)
	group.POST("/chat/session", r.createSession)
	group.DELETE("/chat/session/:session_id", r.closeSession)

	// 系统提示词 CRUD
	group.POST("/system-prompts", r.createSystemPrompt)
	group.GET("/system-prompts", r.listSystemPrompts)
	group.GET("/system-prompts/search", r.searchSystemPromptsByName)
	group.GET("/system-prompts/:id", r.getSystemPrompt)
	group.PUT("/system-prompts/:id", r.updateSystemPrompt)
	group.PATCH("/system-prompts/:id/status", r.updateSystemPromptStatus)
	group.PATCH("/system-prompts/:id/default", r.setSystemPromptAsDefault)
	group.DELETE("/system-prompts/:id", r.deleteSystemPrompt)
	group.DELETE("/system-prompts", r.batchDeleteSystemPrompts)
}
