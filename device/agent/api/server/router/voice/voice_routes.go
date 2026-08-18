package voice

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/YOUR-ORG/sensecraft-voice/device/agent/api/server/httputils"
	appcfg "github.com/YOUR-ORG/sensecraft-voice/device/agent/cmd/app/config"
	pvoice "github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/plugins/voice"
	httpclient "github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/util/http"
	"github.com/gorilla/websocket"
)

type recordReq struct {
	Action     string `json:"action" binding:"required,oneof=start stop"`
	DeviceID   string `json:"deviceId"`
	SampleRate int    `json:"sampleRate"`
	Channels   int    `json:"channels"`
	FilePath   string `json:"filePath"`
	SoftMute   *bool  `json:"softMute"`
	Output     string `json:"output"`     // 可选: file | stream | both
	ManualStop *bool  `json:"manualStop"` // 可选: 是否手动停止，默认为true
}

type quickReq struct {
	Seconds    int    `json:"seconds"`
	DeviceID   string `json:"deviceId"`
	SampleRate int    `json:"sampleRate"`
	Channels   int    `json:"channels"`
	Dir        string `json:"dir"`
}

func (r *Router) record(c *gin.Context) {
	req := &recordReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		resp := httputils.NewResponse()
		httputils.SetFailedWithCode(c, resp, http.StatusBadRequest, err)
		return
	}

	resp := httputils.NewResponse()
	switch req.Action {
	case "start":
		// 构造覆盖项：仅填充传入的字段
		over := appcfg.VoiceOptions{}
		useOverride := false
		if req.DeviceID != "" {
			over.DeviceID = req.DeviceID
			useOverride = true
		}
		if req.SampleRate > 0 {
			over.SampleRate = req.SampleRate
			useOverride = true
		}
		if req.Channels > 0 {
			over.Channels = req.Channels
			useOverride = true
		}
		if req.FilePath != "" {
			over.FilePath = req.FilePath
			useOverride = true
		}
		if req.Output != "" {
			over.Output = req.Output
			useOverride = true
		}

		if useOverride {
			if err := r.c.Voice().StartWithOverride(context.Background(), over); err != nil {
				httputils.SetFailedWithCode(c, resp, http.StatusBadRequest, err)
				return
			}
		} else {
			if err := r.c.Voice().StartByConfig(context.Background()); err != nil {
				httputils.SetFailedWithCode(c, resp, http.StatusBadRequest, err)
				return
			}
		}
		resp.Result = map[string]interface{}{"running": true}
		httputils.SetSuccess(c, resp)
	case "stop":
		// 默认手动停止，除非明确指定manualStop为false
		isManualStop := true
		if req.ManualStop != nil {
			isManualStop = *req.ManualStop
		}

		if err := r.c.Voice().StopWithReason(context.Background(), isManualStop); err != nil {
			httputils.SetFailedWithCode(c, resp, http.StatusBadRequest, err)
			return
		}
		resp.Result = map[string]interface{}{
			"running":     false,
			"manual_stop": isManualStop,
		}
		httputils.SetSuccess(c, resp)
	default:
		httputils.SetFailedWithCode(c, resp, http.StatusBadRequest, errors.New("invalid action"))
	}
}

func (r *Router) status(c *gin.Context) {
	resp := httputils.NewResponse()
	running := pvoice.GetManager().IsRunning()
	manualStop := pvoice.GetManager().IsManualStop()
	resp.Result = map[string]interface{}{
		"running":     running,
		"manual_stop": manualStop,
	}
	httputils.SetSuccess(c, resp)
}

func (r *Router) quick(c *gin.Context) {
	req := &quickReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		resp := httputils.NewResponse()
		httputils.SetFailedWithCode(c, resp, http.StatusBadRequest, err)
		return
	}
	resp := httputils.NewResponse()
	path, err := r.c.Voice().QuickRecord(context.Background(), req.Seconds, req.SampleRate, req.Channels, req.DeviceID, req.Dir)
	if err != nil {
		httputils.SetFailedWithCode(c, resp, http.StatusBadRequest, err)
		return
	}
	resp.Result = map[string]interface{}{"path": path}
	httputils.SetSuccess(c, resp)
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

// asrWS: 将 wsSink 广播出来的消息转发给连接的客户端
func (r *Router) asrWS(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	ch, unsubscribe := pvoice.GetASRHub().Subscribe(32)
	defer unsubscribe()
	defer ws.Close()
	for msg := range ch {
		if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

// deviceStatus: 获取设备状态信息
func (r *Router) deviceStatus(c *gin.Context) {
	resp := httputils.NewResponse()

	// 获取设备状态
	deviceStatus := pvoice.GetManager().GetDeviceStatus()

	// 添加远程服务配置信息
	deviceStatus["remote_base_url"] = r.cfg.Remote.BaseURL

	resp.Result = deviceStatus
	httputils.SetSuccess(c, resp)
}

// asrCacheStatus: 获取ASR缓存状态
func (r *Router) asrCacheStatus(c *gin.Context) {
	resp := httputils.NewResponse()

	// 获取ASR缓存状态
	status := pvoice.GetManager().GetASRCacheStatus()

	resp.Result = status
	httputils.SetSuccess(c, resp)
}

// asrCacheMetrics: 获取ASR缓存指标
func (r *Router) asrCacheMetrics(c *gin.Context) {
	resp := httputils.NewResponse()

	// 获取ASR缓存指标
	metrics := pvoice.GetManager().GetASRCacheMetrics()

	resp.Result = metrics
	httputils.SetSuccess(c, resp)
}

// asrCacheRetry: 强制重试ASR缓存
func (r *Router) asrCacheRetry(c *gin.Context) {
	resp := httputils.NewResponse()

	// 获取ASR缓存Sink
	asrCacheSink := pvoice.GetManager().GetASRCacheSink()
	if asrCacheSink == nil {
		httputils.SetFailedWithCode(c, resp, http.StatusBadRequest, errors.New("ASR cache not initialized"))
		return
	}

	// 执行强制重试
	if err := asrCacheSink.RetryFailedCache(); err != nil {
		httputils.SetFailedWithCode(c, resp, http.StatusInternalServerError, err)
		return
	}

	resp.Result = map[string]interface{}{
		"message":   "ASR cache retry triggered successfully",
		"timestamp": time.Now().UnixMilli(),
	}
	httputils.SetSuccess(c, resp)
}

// asrCacheCleanup: 清理ASR缓存
func (r *Router) asrCacheCleanup(c *gin.Context) {
	resp := httputils.NewResponse()

	// 获取ASR缓存Sink
	asrCacheSink := pvoice.GetManager().GetASRCacheSink()
	if asrCacheSink == nil {
		httputils.SetFailedWithCode(c, resp, http.StatusBadRequest, errors.New("ASR cache not initialized"))
		return
	}

	// 执行清理
	if err := asrCacheSink.CleanupExpiredCache(); err != nil {
		httputils.SetFailedWithCode(c, resp, http.StatusInternalServerError, err)
		return
	}

	resp.Result = map[string]interface{}{
		"message":   "ASR cache cleanup completed successfully",
		"timestamp": time.Now().UnixMilli(),
	}
	httputils.SetSuccess(c, resp)
}

// getRemoteConfig: 获取远程配置
func (r *Router) getRemoteConfig(c *gin.Context) {
	resp := httputils.NewResponse()

	// 获取当前远程配置
	baseURL := r.opt.GetRemoteConfig()

	resp.Result = map[string]interface{}{
		"base_url": baseURL,
		"enabled":  r.cfg.Remote.AudioStream.Enabled,
	}
	httputils.SetSuccess(c, resp)
}

// updateRemoteConfigReq 更新远程配置请求
type updateRemoteConfigReq struct {
	BaseURL string `json:"base_url" binding:"required"`
}

// updateRemoteConfig: 更新远程配置
func (r *Router) updateRemoteConfig(c *gin.Context) {
	req := &updateRemoteConfigReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		resp := httputils.NewResponse()
		httputils.SetFailedWithCode(c, resp, http.StatusBadRequest, err)
		return
	}

	resp := httputils.NewResponse()

	// 更新远程配置
	if err := r.opt.UpdateRemoteConfig(req.BaseURL); err != nil {
		httputils.SetFailedWithCode(c, resp, http.StatusBadRequest, err)
		return
	}

	resp.Result = map[string]interface{}{
		"message":   "Remote config updated successfully",
		"base_url":  req.BaseURL,
		"timestamp": time.Now().UnixMilli(),
	}
	httputils.SetSuccess(c, resp)
}

// testRemoteConnectionReq 测试远程连接请求
type testRemoteConnectionReq struct {
	BaseURL string `json:"base_url" binding:"required"`
}

// testRemoteConnection: 测试远程连接
func (r *Router) testRemoteConnection(c *gin.Context) {
	req := &testRemoteConnectionReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		resp := httputils.NewResponse()
		httputils.SetFailedWithCode(c, resp, http.StatusBadRequest, err)
		return
	}

	resp := httputils.NewResponse()

	// 验证URL格式
	baseURL := strings.TrimSpace(req.BaseURL)
	if baseURL == "" {
		httputils.SetFailedWithCode(c, resp, http.StatusBadRequest, errors.New("base_url is required"))
		return
	}

	// 解析URL
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		httputils.SetFailedWithCode(c, resp, http.StatusBadRequest, errors.New("invalid URL format"))
		return
	}

	// 检查协议
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		httputils.SetFailedWithCode(c, resp, http.StatusBadRequest, errors.New("only http and https protocols are supported"))
		return
	}

	// 构造测试URL - 尝试多个可能的端点
	testURLs := []string{
		baseURL + "/healthz",
		baseURL + "/",
		baseURL + "/health",
	}

	var testResult map[string]interface{}
	var lastErr error

	// 创建HTTP客户端，设置较短的超时时间
	httpClient := httpclient.NewClient().SetTimeout(10 * time.Second)

	// 尝试连接
	for _, testURL := range testURLs {
		startTime := time.Now()
		httpResp, err := httpClient.Get(testURL)
		responseTime := time.Since(startTime)

		if err != nil {
			lastErr = err
			continue
		}

		// 200或404都表示服务可达
		if httpResp.StatusCode == 200 || httpResp.StatusCode == 404 {
			testResult = map[string]interface{}{
				"reachable":        true,
				"status_code":      httpResp.StatusCode,
				"response_time_ms": responseTime.Milliseconds(),
				"test_url":         testURL,
				"message":          "连接成功",
			}
			break
		}

		// 其他状态码也记录，但继续尝试其他URL
		testResult = map[string]interface{}{
			"reachable":        false,
			"status_code":      httpResp.StatusCode,
			"response_time_ms": responseTime.Milliseconds(),
			"test_url":         testURL,
			"message":          "服务响应异常",
		}
	}

	// 如果所有URL都失败
	if testResult == nil {
		testResult = map[string]interface{}{
			"reachable":        false,
			"status_code":      0,
			"response_time_ms": 0,
			"test_url":         "",
			"message":          "连接失败: " + lastErr.Error(),
		}
	}

	resp.Result = testResult
	httputils.SetSuccess(c, resp)
}

// serveRecordingFile: 提供录音文件服务
func (r *Router) serveRecordingFile(c *gin.Context) {
	// 获取文件路径参数
	filePath := c.Param("filepath")
	if filePath == "" {
		c.String(http.StatusBadRequest, "Missing file path")
		return
	}

	// 构造完整的文件路径
	fullPath := "./recordings/" + filePath

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.String(http.StatusNotFound, "File not found")
		return
	}

	// 设置正确的Content-Type
	c.Header("Content-Type", "audio/wav")

	// 提供文件下载
	c.File(fullPath)
}
