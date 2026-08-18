package voice

import (
	"github.com/gin-gonic/gin"

	"github.com/YOUR-ORG/sensecraft-voice/device/agent/cmd/app/config"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/cmd/app/options"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/controller"
)

type Router struct {
	c   controller.SensecraftVoiceInterface
	cfg config.Config
	opt *options.Options
}

func NewRouter(o *options.Options) {
	r := &Router{
		c:   o.Controller,
		cfg: o.ComponentConfig,
		opt: o,
	}
	r.initRoutes(o.HttpEngine)
}

func (r *Router) initRoutes(httpEngine *gin.Engine) {
	grp := httpEngine.Group("/v1/voice")
	{
		grp.POST("/record", r.record)
		grp.GET("/status", r.status)
		grp.POST("/quick", r.quick)
		grp.GET("/asr-ws", r.asrWS)
		grp.GET("/device/status", r.deviceStatus) // 新增设备状态接口

		// ASR缓存相关接口
		grp.GET("/asr-cache/status", r.asrCacheStatus)      // 获取ASR缓存状态
		grp.GET("/asr-cache/metrics", r.asrCacheMetrics)    // 获取ASR缓存指标
		grp.POST("/asr-cache/retry", r.asrCacheRetry)       // 强制重试ASR缓存
		grp.DELETE("/asr-cache/cleanup", r.asrCacheCleanup) // 清理ASR缓存

		// 配置管理接口
		grp.GET("/config/remote", r.getRemoteConfig)            // 获取远程配置
		grp.PUT("/config/remote", r.updateRemoteConfig)         // 更新远程配置
		grp.POST("/config/remote/test", r.testRemoteConnection) // 测试远程连接
	}

	// 录音文件服务 - 在voice路由组中
	httpEngine.GET("/recordings/*filepath", r.serveRecordingFile)
}
