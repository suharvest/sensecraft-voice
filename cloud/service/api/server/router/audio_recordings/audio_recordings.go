package audio_recordings

import (
	"github.com/gin-gonic/gin"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/middleware"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/options"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller"
)

type audioRecordingsRouter struct {
	c controller.SensecraftVoiceInterface
}

func NewRouter(o *options.Options) {
	router := &audioRecordingsRouter{
		c: o.Controller,
	}
	router.initRoutes(o, o.HttpEngine)
}

func (a *audioRecordingsRouter) initRoutes(o *options.Options, httpEngine *gin.Engine) {
	recordingsRoute := httpEngine.Group("/api/v1/audio")
	{
		// 音频录音上传（设备侧，按单条路由挂设备鉴权）
		recordingsRoute.POST("/upload", middleware.DeviceAuth(o), a.uploadAudioRecording)

		// 音频录音管理
		recordingsRoute.GET("/", a.listAudioRecordings)
		recordingsRoute.GET("/:id", a.getAudioRecording)

		// 直接播放音频（通过session_id和audio_id）- 必须在更具体的路由之前
		recordingsRoute.GET("/session/:session_id/audio/:audio_id/play", a.playAudioRecording)
		recordingsRoute.GET("/session/:session_id/audio/:audio_id", a.getAudioRecordingBySessionAndAudio)

		recordingsRoute.DELETE("/:id", a.deleteAudioRecording)

		// 音频录音下载
		recordingsRoute.GET("/:id/download", a.downloadAudioRecording)
	}
}
